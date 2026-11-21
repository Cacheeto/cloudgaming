package ar.cloudgaming.app;

import android.os.SystemClock;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Set;
import java.util.concurrent.Callable;
import java.util.concurrent.ExecutionException;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.Future;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TimeoutException;

/**
 * Busca las PC con CloudGaming en la red del celu: le pide a la red la lista de equipos y
 * le pregunta a cada uno, todos a la vez, si tiene el panel (GET /api/hola).
 *
 * Así la app no trae la dirección de ninguna PC: le sirve a cualquiera con su propia cuenta.
 */
final class BuscarPc {
    /** Cuánto se espera, como mucho, a que contesten. */
    private static final int ESPERA_MS = 4000;
    /** Los celulares y tablets no tienen el panel: ni se les pregunta. */
    private static final Set<String> SIN_PANEL = new HashSet<>(Arrays.asList("android", "iOS"));

    static final class Pc {
        final String ip;
        final String nombre;
        /** De la misma cuenta que el celu. Si no, es un equipo que otra cuenta le compartió. */
        final boolean deMiCuenta;

        Pc(String ip, String nombre, boolean deMiCuenta) {
            this.ip = ip;
            this.nombre = nombre;
            this.deMiCuenta = deMiCuenta;
        }
    }

    static final class Resultado {
        /** Primero las de la cuenta del celu. */
        final List<Pc> pcs;
        /** Con qué cuenta está el celu en la red (p. ej. el correo), o null. */
        final String cuenta;

        Resultado(List<Pc> pcs, String cuenta) {
            this.pcs = pcs;
            this.cuenta = cuenta;
        }

        Pc buscar(String ip) {
            for (Pc pc : pcs) {
                if (pc.ip.equals(ip)) {
                    return pc;
                }
            }
            return null;
        }
    }

    private BuscarPc() {
    }

    /** Hace la búsqueda. Tarda unos segundos: nunca desde el hilo principal. */
    static Resultado buscar(Red red) throws Exception {
        JSONObject estado = red.estadoDeLaRed();
        long yo = estado.getJSONObject("Self").optLong("UserID");
        JSONObject usuarios = estado.optJSONObject("User");
        JSONObject usuario = usuarios != null ? usuarios.optJSONObject(String.valueOf(yo)) : null;
        String cuenta = usuario != null ? usuario.optString("LoginName", null) : null;

        List<Pc> pcs = new ArrayList<>();
        JSONObject pares = estado.optJSONObject("Peer");
        if (pares == null) {
            return new Resultado(pcs, cuenta); // el celu está solo en la red
        }
        ExecutorService hilos = Executors.newCachedThreadPool();
        try {
            List<Future<Pc>> pedidos = new ArrayList<>();
            for (Iterator<String> claves = pares.keys(); claves.hasNext(); ) {
                JSONObject par = pares.getJSONObject(claves.next());
                final String ip = ipv4(par.optJSONArray("TailscaleIPs"));
                if (ip == null || SIN_PANEL.contains(par.optString("OS"))) {
                    continue;
                }
                final String nombreEnLaRed = par.optString("HostName", ip);
                final boolean mia = par.optLong("UserID") == yo;
                pedidos.add(hilos.submit(new Callable<Pc>() {
                    @Override
                    public Pc call() {
                        String nombre = preguntar(ip);
                        if (nombre == null) {
                            return null;
                        }
                        return new Pc(ip, nombre.isEmpty() ? nombreEnLaRed : nombre, mia);
                    }
                }));
            }
            long fin = SystemClock.elapsedRealtime() + ESPERA_MS;
            for (Future<Pc> pedido : pedidos) {
                try {
                    Pc pc = pedido.get(Math.max(0, fin - SystemClock.elapsedRealtime()), TimeUnit.MILLISECONDS);
                    if (pc != null) {
                        pcs.add(pc);
                    }
                } catch (ExecutionException | TimeoutException e) {
                    // esa no contestó a tiempo: apagada, o no es una PC con CloudGaming
                }
            }
        } finally {
            hilos.shutdownNow();
        }
        Collections.sort(pcs, new Comparator<Pc>() {
            @Override
            public int compare(Pc a, Pc b) {
                return Boolean.compare(b.deMiCuenta, a.deMiCuenta);
            }
        });
        return new Resultado(pcs, cuenta);
    }

    private static String ipv4(JSONArray ips) {
        for (int i = 0; ips != null && i < ips.length(); i++) {
            String ip = ips.optString(i);
            if (ip.indexOf('.') > 0) {
                return ip;
            }
        }
        return null;
    }

    /** El nombre que da el panel de esa PC ("" si no da ninguno), o null si ahí no hay panel. */
    private static String preguntar(String ip) {
        JSONObject dato = hola(ip);
        return dato != null && dato.optInt("cloudgaming") == 1 ? dato.optString("nombre", "") : null;
    }

    /** Lo que contesta el panel en GET /api/hola (nombre, versión, teclado), o null si no contesta. */
    static JSONObject hola(String ip) {
        HttpURLConnection conexion = null;
        try {
            conexion = (HttpURLConnection) new URL(MiPc.panel(ip) + "api/hola").openConnection();
            conexion.setConnectTimeout(ESPERA_MS);
            conexion.setReadTimeout(ESPERA_MS);
            if (conexion.getResponseCode() != HttpURLConnection.HTTP_OK) {
                return null;
            }
            try (InputStream entrada = conexion.getInputStream()) {
                ByteArrayOutputStream datos = new ByteArrayOutputStream();
                byte[] bloque = new byte[4096];
                int leidos;
                while ((leidos = entrada.read(bloque)) != -1 && datos.size() < 65536) {
                    datos.write(bloque, 0, leidos);
                }
                return new JSONObject(datos.toString("UTF-8"));
            }
        } catch (IOException | JSONException e) {
            return null;
        } finally {
            if (conexion != null) {
                conexion.disconnect();
            }
        }
    }
}
