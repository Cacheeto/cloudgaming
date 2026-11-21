package ar.cloudgaming.app;

import android.app.Activity;
import android.app.Application;
import android.content.Context;
import android.content.Intent;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.VpnService;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.os.SystemClock;
import android.util.Log;

import org.json.JSONObject;

import java.io.File;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import java.util.concurrent.Callable;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;

import libtailscale.Libtailscale;

/**
 * La red propia de la app: el motor de Tailscale (libtailscale) corriendo acá adentro,
 * sin la app de Tailscale. Con él, el celu llega a la PC desde cualquier lado.
 *
 * Esta clase arranca el motor, sigue su estado y decide cuándo armar el túnel
 * (ServicioRed): mientras la app está a la vista, y unos minutos más.
 */
final class Red implements VigiaRed.Oyente, libtailscale.NotificationCallback,
        Application.ActivityLifecycleCallbacks {
    // Lo que le falta a la red para andar (o nada), para el menú.
    /** Android todavía no dio permiso para armar el túnel (la «VPN»). */
    static final String PERMISO = "permiso";
    /** Este celu todavía no inició sesión en la red. */
    static final String SESION = "sesion";
    /** Inició sesión, pero la red pide que un administrador apruebe el equipo. */
    static final String APROBACION = "aprobacion";
    static final String CONECTANDO = "conectando";
    static final String LISTA = "lista";

    private static final String TAG = "Red";
    // Estados del motor (ipn.State de Tailscale).
    private static final int NECESITA_SESION = 2, NECESITA_APROBACION = 3, DETENIDA = 4, EN_MARCHA = 6;
    // Qué avisos pedirle (ipn.NotifyWatchOpt): el estado inicial, y que no mande ni las
    // claves privadas ni el mapa completo de la red.
    private static final long AVISOS = 2 | 16 | 8192;
    // Con la app fuera de la vista, cuánto se mantiene el túnel antes de bajarlo.
    private static final long APAGAR_A_LOS_MS = 5 * 60 * 1000;
    // Si el túnel se cae, cuánto esperar antes de armarlo de nuevo: que un túnel que no
    // logra armarse no quede en un ida y vuelta sin pausa.
    private static final long REINTENTAR_A_LOS_MS = 10 * 1000;
    // Iniciar sesión se hace en el navegador, con la app fuera de la vista: mientras tanto
    // el servicio no se baja (sin él, Android le corta la red a la app y el motor no se
    // entera de que la sesión terminó). Con tope, por si el inicio de sesión se abandona.
    private static final long SESION_HASTA_MS = 30 * 60 * 1000;

    private static Red unica;

    private final Application app;
    private final Handler principal = new Handler(Looper.getMainLooper());
    // Las llamadas al motor pueden demorar (esperan a que termine de arrancar): van en fila
    // por un hilo aparte, nunca por el principal.
    private final ExecutorService fila = Executors.newSingleThreadExecutor();
    private final VigiaRed vigia = new VigiaRed(this);
    private final List<Runnable> oyentes = new CopyOnWriteArrayList<>();
    private volatile libtailscale.Application motor;

    private volatile int estadoMotor;
    private volatile boolean tunel;
    private volatile ServicioRed servicio;
    // Solo desde el hilo principal:
    private boolean tunelPedido;
    private long noAntesDe;
    private long sesionHasta;
    private final Set<Activity> aLaVista = new HashSet<>();
    // El usuario tocó «Iniciar sesión» y falta abrirle la página (ver urlDeSesion).
    private volatile boolean esperaSesion;
    private volatile String urlSesion;

    static synchronized Red de(Context contexto) {
        if (unica == null) {
            unica = new Red((Application) contexto.getApplicationContext());
        }
        return unica;
    }

    private Red(Application app) {
        this.app = app;
        app.registerActivityLifecycleCallbacks(this);
        fila.execute(new Runnable() {
            @Override
            public void run() {
                File carpeta = new File(app.getFilesDir(), "red");
                carpeta.mkdirs();
                motor = Libtailscale.start(carpeta.getPath(), carpeta.getPath(), false,
                    new EntornoRed(app, vigia));
                vigia.empezar((ConnectivityManager) app.getSystemService(Context.CONNECTIVITY_SERVICE));
                motor.watchNotifications(AVISOS, Red.this);
            }
        });
    }

    // ------------------------------------------------------------ para el menú

    /** Qué le falta a la red (PERMISO, SESION, APROBACION), o CONECTANDO, o LISTA. */
    String estado() {
        if (VpnService.prepare(app) != null) {
            return PERMISO;
        }
        int estado = estadoMotor;
        if (estado == NECESITA_SESION) {
            return SESION;
        }
        if (estado == NECESITA_APROBACION) {
            return APROBACION;
        }
        return estado == EN_MARCHA && tunel ? LISTA : CONECTANDO;
    }

    /** Avisa (en el hilo principal) cada vez que puede haber cambiado estado(). */
    void escuchar(Runnable oyente) {
        oyentes.add(oyente);
    }

    void dejarDeEscuchar(Runnable oyente) {
        oyentes.remove(oyente);
    }

    /** Espera a que la red esté lista; para lo que arranca sin pasar por el menú. */
    boolean esperarLista(long ms) throws InterruptedException {
        long fin = SystemClock.elapsedRealtime() + ms;
        synchronized (this) {
            while (!LISTA.equals(estado())) {
                long falta = fin - SystemClock.elapsedRealtime();
                if (falta <= 0) {
                    return false;
                }
                wait(falta);
            }
        }
        return true;
    }

    /**
     * Arma el túnel, si ya se puede. Si falta algo que depende del usuario no hace nada:
     * el menú lo muestra según estado() y vuelve a llamar cuando se resuelve.
     */
    void conectar() {
        if (tunelPedido || estadoMotor < DETENIDA || SystemClock.elapsedRealtime() < noAntesDe
                || VpnService.prepare(app) != null) {
            return;
        }
        tunelPedido = arrancarServicio(ServicioRed.CONECTAR);
    }

    /**
     * Suma este celu a la red: el motor pide una página de inicio de sesión, que llega
     * por urlDeSesion() para abrirla en el navegador.
     */
    void iniciarSesion() {
        esperaSesion = true;
        sesionHasta = SystemClock.elapsedRealtime() + SESION_HASTA_MS;
        arrancarServicio(ServicioRed.SOLO_AVISO);
        fila.execute(new Runnable() {
            @Override
            public void run() {
                try {
                    // Los mismos tres pasos que la app de Tailscale: las preferencias de
                    // ahora con «quiero estar conectado», reiniciar con ellas, y pedir la página.
                    JSONObject prefs = new JSONObject(api("PATCH", "prefs", "{}"));
                    prefs.put("WantRunning", true);
                    api("POST", "start", new JSONObject().put("UpdatePrefs", prefs).toString());
                    api("POST", "login-interactive", null);
                } catch (Exception e) {
                    Log.e(TAG, "no se pudo iniciar sesión", e);
                    esperaSesion = false;
                }
            }
        });
    }

    /** La página para iniciar sesión, una sola vez por pedido (después devuelve null). */
    String urlDeSesion() {
        String url = urlSesion;
        urlSesion = null;
        return url;
    }

    /** Saca este celu de la red, para entrar con otra cuenta (vuelve a pedir iniciar sesión). */
    void cerrarSesion() {
        fila.execute(new Runnable() {
            @Override
            public void run() {
                try {
                    api("POST", "logout", null);
                } catch (Exception e) {
                    Log.e(TAG, "no se pudo cerrar la sesión", e);
                }
            }
        });
    }

    /**
     * Los equipos de la red y la cuenta de este celu (el «status» de Tailscale, el mismo
     * de `tailscale status --json`). Espera al motor: nunca desde el hilo principal.
     */
    JSONObject estadoDeLaRed() throws Exception {
        return new JSONObject(fila.submit(new Callable<String>() {
            @Override
            public String call() throws Exception {
                return api("GET", "status", null);
            }
        }).get(30, TimeUnit.SECONDS));
    }

    // ------------------------------------------------------ para ServicioRed

    void servicioCreado(ServicioRed nuevo) {
        servicio = nuevo;
    }

    /** Le pasa el servicio al motor para que arme el túnel sobre él. */
    void pedirTunel(final ServicioRed para) {
        fila.execute(new Runnable() {
            @Override
            public void run() {
                quiereConexion(true);
                Libtailscale.requestVPN(para);
            }
        });
    }

    void tunelArmado(boolean armado) {
        tunel = armado;
        cambio();
    }

    void servicioCerrado(final ServicioRed cual) {
        if (servicio == cual) {
            servicio = null;
            tunel = false;
            tunelPedido = false;
            noAntesDe = SystemClock.elapsedRealtime() + REINTENTAR_A_LOS_MS;
            principal.postDelayed(alCambiar, REINTENTAR_A_LOS_MS);
        }
        fila.execute(new Runnable() {
            @Override
            public void run() {
                Libtailscale.serviceDisconnect(cual);
            }
        });
        cambio();
    }

    /** Que el motor se desconecte de la red (no pierde la sesión). */
    void apagarMotor() {
        fila.execute(new Runnable() {
            @Override
            public void run() {
                quiereConexion(false);
            }
        });
    }

    Network redDeSalida() {
        return vigia.actual();
    }

    @Override
    public void cambioLaRed(Network red) {
        ServicioRed s = servicio;
        if (s != null) {
            s.salirPor(red);
        }
    }

    // ------------------------------------------------------------- con el motor

    /** Aviso del motor (un ipn.Notify en JSON): cambió su estado o hay algo para mostrar. */
    @Override
    public void onNotify(byte[] datos) throws Exception {
        JSONObject aviso = new JSONObject(new String(datos, StandardCharsets.UTF_8));
        if (!aviso.isNull("State")) {
            estadoMotor = aviso.getInt("State");
        }
        String url = aviso.isNull("BrowseToURL") ? null : aviso.getString("BrowseToURL");
        if (url != null && url.startsWith("https://") && esperaSesion) {
            esperaSesion = false;
            urlSesion = url;
        }
        if (!aviso.isNull("ErrMessage")) {
            Log.w(TAG, "el motor avisa: " + aviso.getString("ErrMessage"));
        }
        cambio();
    }

    private void quiereConexion(boolean si) {
        try {
            api("PATCH", "prefs", new JSONObject().put("WantRunning", si).put("WantRunningSet", true).toString());
        } catch (Exception e) {
            Log.e(TAG, "no se pudo cambiar WantRunning a " + si, e);
        }
    }

    /** Llamada a la API local de Tailscale (la misma que usa su app). Solo desde la fila. */
    private String api(String metodo, String ruta, String cuerpo) throws Exception {
        libtailscale.LocalAPIResponse respuesta = motor.callLocalAPI(30000, metodo, "/localapi/v0/" + ruta,
            cuerpo != null ? new Cuerpo(cuerpo) : null);
        byte[] datos = respuesta.bodyBytes();
        String texto = datos != null ? new String(datos, StandardCharsets.UTF_8) : "";
        if (respuesta.statusCode() >= 400) {
            throw new IOException(metodo + " " + ruta + ": " + respuesta.statusCode() + " " + texto);
        }
        return texto;
    }

    /** El cuerpo de un pedido, como lo lee el motor: de a pedazos, y null al terminar. */
    private static final class Cuerpo implements libtailscale.InputStream {
        private byte[] datos;

        Cuerpo(String texto) {
            datos = texto.getBytes(StandardCharsets.UTF_8);
        }

        @Override
        public byte[] read() {
            byte[] todo = datos;
            datos = null;
            return todo;
        }

        @Override
        public void close() {
        }
    }

    // ---------------------------------------------------------------- el ritmo

    private final Runnable alCambiar = new Runnable() {
        @Override
        public void run() {
            // Con la app a la vista, o a mitad de un inicio de sesión (el servicio ya anda).
            if (!aLaVista.isEmpty() || servicio != null) {
                conectar();
            }
            for (Runnable oyente : oyentes) {
                oyente.run();
            }
        }
    };

    private final Runnable apagar = new Runnable() {
        @Override
        public void run() {
            ServicioRed s = servicio;
            if (!aLaVista.isEmpty() || s == null) {
                return;
            }
            long faltaSesion = sesionHasta - SystemClock.elapsedRealtime();
            if (estadoMotor == NECESITA_SESION && faltaSesion > 0) {
                principal.postDelayed(this, faltaSesion); // inicio de sesión en curso
                return;
            }
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q && s.isAlwaysOn()) {
                return; // el usuario eligió «VPN siempre activa»: se queda
            }
            apagarMotor();
            s.close();
        }
    };

    private void cambio() {
        synchronized (this) {
            notifyAll();
        }
        principal.post(alCambiar);
    }

    private boolean arrancarServicio(String accion) {
        try {
            app.startService(new Intent(app, ServicioRed.class).setAction(accion));
            return true;
        } catch (IllegalStateException e) {
            // Android no deja arrancar servicios con la app en segundo plano: será al volver.
            return false;
        }
    }

    @Override
    public void onActivityStarted(Activity actividad) {
        aLaVista.add(actividad);
        principal.removeCallbacks(apagar);
        conectar();
    }

    @Override
    public void onActivityStopped(Activity actividad) {
        if (aLaVista.remove(actividad) && aLaVista.isEmpty()) {
            principal.postDelayed(apagar, APAGAR_A_LOS_MS);
        }
    }

    @Override
    public void onActivityCreated(Activity actividad, Bundle guardado) {
    }

    @Override
    public void onActivityResumed(Activity actividad) {
    }

    @Override
    public void onActivityPaused(Activity actividad) {
    }

    @Override
    public void onActivitySaveInstanceState(Activity actividad, Bundle salida) {
    }

    @Override
    public void onActivityDestroyed(Activity actividad) {
    }
}
