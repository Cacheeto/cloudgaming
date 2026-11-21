package ar.cloudgaming.app;

import android.app.Activity;
import android.app.UiModeManager;
import android.content.ActivityNotFoundException;
import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.content.res.Configuration;
import android.graphics.Color;
import android.net.Uri;
import android.net.VpnService;
import android.os.Build;
import android.os.Bundle;
import android.util.Log;
import android.view.InputDevice;
import android.view.KeyEvent;
import android.view.MotionEvent;
import android.view.View;
import android.view.WindowManager;
import android.webkit.CookieManager;
import android.webkit.DownloadListener;
import android.webkit.JavascriptInterface;
import android.webkit.ValueCallback;
import android.webkit.WebResourceError;
import android.webkit.WebResourceRequest;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.Toast;

import com.limelight.BuildConfig;

import org.json.JSONException;

/**
 * Pantalla principal: el panel de la PC a pantalla completa.
 *
 * El joystick se traduce a teclas (flechas, Enter, Escape) que el panel ya entiende,
 * y la página puede pedir "jugar" por el objeto JavaScript CloudGaming: el stream
 * (el código de Moonlight) corre dentro de esta misma app.
 *
 * A la PC se llega por la red propia de la app (ver Red). Mientras a esa red le falte
 * algo que depende del usuario, en lugar del panel se muestra una página que lo pide.
 * La primera vez esas páginas son un asistente de tres pasos: el permiso, la cuenta de
 * la red y elegir la PC (que la app busca sola en la red, ver BuscarPc y MiPc).
 */
public class Menu extends Activity {
    static final String VERSION = BuildConfig.VERSION_NAME;

    private static final int PEDIDO_VPN = 1;
    /** La red anda pero el panel no carga, y el celu tampoco tiene la copia guardada. */
    private static final String SIN_CONEXION = "sin_conexion";
    /** Falta iniciar sesión y no hay navegador (la tele): un QR para hacerlo desde el celu. */
    private static final String CODIGO = "codigo";
    // Todavía no hay PC elegida: se busca en la red, y se elige si hay más de una.
    private static final String BUSCANDO = "buscando";
    private static final String ELEGIR = "elegir";
    private static final String SIN_PC = "sin_pc";

    private WebView web;
    private Red red;
    /**
     * Qué página propia se está mostrando en lugar del panel: SIN_CONEXION, CODIGO, una de
     * la búsqueda de la PC, o lo que le falta a la red (un estado de Red). null: el panel.
     */
    private String aviso;
    /** La página de inicio de sesión de la red, cuando hay que mostrarla como QR. */
    private String sesionEnCodigo;
    private boolean tele;
    private boolean buscando;
    /** Lo que encontró la última búsqueda de la PC (null si falló). */
    private BuscarPc.Resultado hallado;
    /** Vino de «Cambiar de PC»: mostrar la lista aunque haya una sola. */
    private boolean elegirAMano;
    private final Runnable alCambiarLaRed = new Runnable() {
        @Override
        public void run() {
            acomodar();
        }
    };

    @Override
    protected void onCreate(Bundle guardado) {
        super.onCreate(guardado);
        usarZonaCamara(this);
        tele = esTele(this);
        red = Red.de(this);
        web = new WebView(this);
        web.setBackgroundColor(Color.rgb(4, 10, 28));
        WebSettings ajustes = web.getSettings();
        ajustes.setJavaScriptEnabled(true);
        ajustes.setDomStorageEnabled(true);
        ajustes.setMediaPlaybackRequiresUserGesture(false); // la música del menú arranca sola
        ajustes.setUserAgentString(ajustes.getUserAgentString() + " CloudGamingApp/" + VERSION);
        CookieManager.getInstance().setAcceptCookie(true);
        web.addJavascriptInterface(new ApiJs(), "CloudGaming");
        web.setWebViewClient(new Cliente());
        web.setDownloadListener(new DownloadListener() {
            @Override
            public void onDownloadStart(String url, String agente, String disposicion, String tipo, long largo) {
                if (Actualizador.esApk(url, tipo)) {
                    Actualizador.actualizar(Menu.this, url); // por la red de la app: el navegador no llega
                } else {
                    abrirAfuera(Uri.parse(url));
                }
            }
        });
        setContentView(web);
        if (!Red.LISTA.equals(red.estado())) {
            // El panel va a la IP de Tailscale de la PC: pedido antes de que el túnel esté
            // listo, sale por la red común y queda colgado unos minutos (fondo azul, sin
            // nada). Mientras tanto, la página de la red; acomodar lo carga cuando esté lista.
            if (MiPc.panel(this) != null) {
                mostrarAviso(red.estado());
            }
        } else if (guardado == null || web.restoreState(guardado) == null) {
            String panel = MiPc.panel(this);
            if (panel != null) {
                web.loadUrl(panel);
            }
        }
        // Sin PC elegida todavía: el asistente arranca en onStart (acomodar).
    }

    // ------------------------------------------------------------ navegación web

    // Páginas propias, para cuando no se puede mostrar el panel.
    // Con el joystick: flechas entre botones y A para tocar.
    private String paginaDe(String aviso) {
        String equipo = tele ? "esta tele" : "este celu";
        // La primera vez (sin PC elegida) las páginas de la red son los pasos del asistente.
        boolean primeraVez = MiPc.ip(this) == null;
        String reintentar = primeraVez ? "cloudgaming://pc/buscar" : MiPc.panel(this);
        switch (aviso) {
            case Red.PERMISO:
                return pagina(primeraVez ? "Paso 1 de 3" : null,
                    primeraVez ? "Te damos la bienvenida a CloudGaming" : "Falta un permiso",
                    (primeraVez ? "Vamos a dejar todo listo en tres pasos: un permiso, tu cuenta y tu PC.<br>" : "")
                        + "CloudGaming arma una conexión privada con tu PC, y Android pide tu permiso"
                        + " porque la considera una VPN.<br>Solo la usa esta app: el resto "
                        + (tele ? "de la tele" : "del celu") + " sigue igual.",
                    boton("cloudgaming://red/permiso", "Permitir")
                        + "<small>Si usabas la app de Tailscale con «VPN siempre activa», sacale esa opción"
                        + " primero: Android permite una sola VPN a la vez.</small>");
            case Red.SESION:
                return pagina(primeraVez ? "Paso 2 de 3" : null, "Sumá " + equipo + " a tu red",
                    "Entrá con la misma cuenta de Tailscale que usa la PC. Es una sola vez:<br>"
                        + (tele ? "vas a ver un código para escanear con el celu."
                                : "se abre el navegador y, cuando termines, volvé a esta app.")
                        + (primeraVez ? "<br>¿No tenés cuenta? Se crea ahí mismo, con Google, Microsoft o Apple." : ""),
                    boton("cloudgaming://red/sesion", "Iniciar sesión"));
            case CODIGO:
                return pagina(primeraVez ? "Paso 2 de 3" : null, "Escaneá el código con el celu",
                    "<span class='qr'>" + CodigoQr.svg(sesionEnCodigo) + "</span>"
                        + "Entrá con la misma cuenta de Tailscale que usa la PC.<br>"
                        + "Sin cámara: abrí <b>" + sesionEnCodigo + "</b> en cualquier navegador.<br>"
                        + "Cuando termines, " + equipo + " se conecta sola.",
                    boton("cloudgaming://red/sesion", "Pedir otro código"));
            case Red.APROBACION:
                return pagina(primeraVez ? "Paso 2 de 3" : null, "Falta aprobar " + equipo,
                    "Tu red de Tailscale pide que apruebes cada equipo nuevo.<br>"
                        + "Aprobá " + equipo + " en la consola de Tailscale (Machines) y el menú se carga solo.",
                    boton(reintentar, "Reintentar"));
            case BUSCANDO:
                return pagina("Paso 3 de 3", "Buscando tu PC…",
                    "Busco las PC con CloudGaming en tu red.", "");
            case ELEGIR:
                StringBuilder botones = new StringBuilder();
                for (BuscarPc.Pc pc : hallado.pcs) {
                    Uri elegir = new Uri.Builder().scheme("cloudgaming").authority("pc").appendPath("elegir")
                        .appendQueryParameter("ip", pc.ip).build();
                    botones.append(boton(elegir.toString(), escapar(pc.nombre) + (pc.deMiCuenta ? "" : " · de otra cuenta")));
                }
                return pagina("Paso 3 de 3", "¿Con qué PC jugás?",
                    "Encontré estas PC con CloudGaming en tu red:",
                    botones + "<br>" + boton("cloudgaming://pc/buscar", "Buscar de nuevo"));
            case SIN_PC:
                String cuenta = hallado != null && hallado.cuenta != null
                    ? " (<b>" + escapar(hallado.cuenta) + "</b>)" : "";
                return pagina("Paso 3 de 3", "No encontré tu PC",
                    "En la PC tiene que estar instalado CloudGaming Host, con la misma cuenta de la red que "
                        + equipo + cuenta + ".<br>Revisá también que la PC esté prendida.",
                    boton("cloudgaming://pc/buscar", "Buscar de nuevo")
                        + boton("cloudgaming://red/salir", "Usar otra cuenta"));
            case SIN_CONEXION:
                return pagina(null, "No llego a la PC",
                    "Revisá que la PC esté prendida y con internet.",
                    boton(reintentar, "Reintentar") + boton("cloudgaming://pc/buscar", "Buscar otra PC"));
            default:
                return pagina(primeraVez ? "Paso 3 de 3" : null, "Conectando con tu PC…",
                    "Si tarda, revisá que " + (tele ? "la tele" : "el celu") + " tenga internet.",
                    boton(reintentar, "Reintentar"));
        }
    }

    private static String pagina(String paso, String titulo, String texto, String botones) {
        return "<!doctype html><html lang='es'><head><meta charset='utf-8'>"
            + "<meta name='viewport' content='width=device-width,initial-scale=1'>"
            + "<style>body{margin:0;height:100vh;display:grid;place-items:center;text-align:center;"
            + "background:#040a1c;color:#f3f6ff;font-family:sans-serif;padding:24px;box-sizing:border-box}"
            + "p{color:#aab4d4;line-height:1.5}small{display:block;margin-top:22px;color:#7d88aa}"
            + ".paso{display:block;color:#7f9cff;font-size:13px;letter-spacing:.12em;text-transform:uppercase}"
            + "h2{margin-top:8px}"
            + "a{display:inline-block;margin:12px 6px 0;padding:12px 22px;border-radius:12px;"
            + "border:1px solid #7f9cff;color:#fff;text-decoration:none;outline:none}"
            + "a:focus{background:#7f9cff33;box-shadow:0 0 0 3px #7f9cff66}"
            + ".qr{display:block;margin:0 auto 18px;width:min(46vh,70vw)}"
            + ".qr svg{display:block;width:100%;border-radius:8px}b{color:#f3f6ff}</style></head>"
            + "<body><div>" + (paso != null ? "<span class='paso'>" + paso + "</span>" : "")
            + "<h2>" + titulo + "</h2><p>" + texto + "</p>" + botones + "</div>"
            + "<script>const b=[...document.querySelectorAll('a')];b[0]&&b[0].focus();"
            + "addEventListener('keydown',e=>{const d={ArrowRight:1,ArrowDown:1,ArrowLeft:-1,ArrowUp:-1}[e.key];"
            + "if(d&&b.length){e.preventDefault();b[(b.indexOf(document.activeElement)+d+b.length)%b.length].focus();}});"
            + "</script></body></html>";
    }

    private static String boton(String destino, String texto) {
        return "<a href='" + escapar(destino) + "'>" + texto + "</a>";
    }

    private static String escapar(String texto) {
        return texto.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")
            .replace("'", "&#39;").replace("\"", "&quot;");
    }

    private class Cliente extends WebViewClient {
        @Override
        public boolean shouldOverrideUrlLoading(WebView vista, WebResourceRequest pedido) {
            Uri url = pedido.getUrl();
            if ("cloudgaming".equals(url.getScheme())) {
                if ("red".equals(url.getAuthority())) {
                    resolver(url.getLastPathSegment());
                } else if ("pc".equals(url.getAuthority())) {
                    resolverPc(url);
                } else {
                    Lanzador.lanzar(Menu.this, url);
                }
                return true;
            }
            String panel = MiPc.panel(Menu.this);
            if (panel != null && url.toString().startsWith(panel)) {
                aviso = null; // «Reintentar» de una página propia
                return false;
            }
            abrirAfuera(url);
            return true;
        }

        @Override
        public void onReceivedError(WebView vista, WebResourceRequest pedido, WebResourceError error) {
            if (pedido.isForMainFrame()) {
                String estado = red.estado();
                mostrarAviso(Red.LISTA.equals(estado) ? SIN_CONEXION : estado);
            }
        }
    }

    /**
     * Decide qué mostrar según la red: la página de lo que le falta, o el panel.
     * Se llama al volver a la app y cada vez que la red cambia de estado.
     */
    private void acomodar() {
        String sesion = red.urlDeSesion();
        if (sesion != null && (tele || !abrirAfuera(Uri.parse(sesion)))) {
            sesionEnCodigo = sesion; // sin navegador: se abre desde el celu
        }
        String estado = red.estado();
        if (!Red.SESION.equals(estado)) {
            sesionEnCodigo = null;
        }
        String panel = MiPc.panel(this);
        if (sesionEnCodigo != null) {
            mostrarAviso(CODIGO);
        } else if (Red.PERMISO.equals(estado) || Red.SESION.equals(estado) || Red.APROBACION.equals(estado)) {
            mostrarAviso(estado);
        } else if (panel == null) {
            // Falta elegir la PC: se busca con la red lista. Lo que muestre la búsqueda queda
            // hasta que se elija una (o se toque «Buscar de nuevo»).
            if (!deLaBusqueda(aviso)) {
                if (Red.LISTA.equals(estado)) {
                    buscarPc();
                } else {
                    mostrarAviso(estado);
                }
            }
        } else if (aviso != null && Red.LISTA.equals(estado)) {
            if (!SIN_CONEXION.equals(aviso)) {
                aviso = null;
                web.loadUrl(panel);
            }
        } else if (aviso != null && !SIN_CONEXION.equals(aviso)) {
            mostrarAviso(estado); // p. ej. de «Iniciar sesión» a «Conectando…»
        }
        // Sin aviso y conectando: se deja el panel, que con su copia guardada se recupera solo.
    }

    private static boolean deLaBusqueda(String aviso) {
        return BUSCANDO.equals(aviso) || ELEGIR.equals(aviso) || SIN_PC.equals(aviso);
    }

    private void mostrarAviso(String cual) {
        if (cual.equals(aviso)) {
            return;
        }
        aviso = cual;
        web.loadDataWithBaseURL(null, paginaDe(cual), "text/html", "utf-8", null);
    }

    /** Los botones de las páginas de la red: cloudgaming://red/permiso, .../sesion y .../salir. */
    private void resolver(String paso) {
        if ("permiso".equals(paso)) {
            Intent pedido = VpnService.prepare(this);
            if (pedido != null) {
                startActivityForResult(pedido, PEDIDO_VPN); // el cartel de Android
                return;
            }
        } else if ("sesion".equals(paso)) {
            sesionEnCodigo = null; // «Pedir otro código»
            red.iniciarSesion();
            Toast.makeText(this, tele ? "Pidiendo el código…" : "Abriendo el navegador…", Toast.LENGTH_LONG).show();
        } else if ("salir".equals(paso)) {
            // «Usar otra cuenta»: fuera de la red; el motor avisa que falta iniciar sesión.
            red.cerrarSesion();
            MiPc.olvidar(this);
            aviso = null;
            mostrarAviso(Red.CONECTANDO);
            return;
        }
        acomodar();
    }

    // ------------------------------------------------------------ elegir la PC

    /** Los botones de la búsqueda: cloudgaming://pc/buscar y .../elegir?ip=... */
    private void resolverPc(Uri url) {
        String paso = url.getLastPathSegment();
        if ("buscar".equals(paso)) {
            cambiarPc();
        } else if ("elegir".equals(paso) && hallado != null) {
            BuscarPc.Pc pc = hallado.buscar(url.getQueryParameter("ip")); // solo una de la lista
            if (pc != null) {
                elegir(pc);
            }
        }
    }

    /** Olvida la PC elegida y la busca de nuevo, mostrando siempre la lista. */
    private void cambiarPc() {
        MiPc.olvidar(this);
        elegirAMano = true;
        aviso = null;
        acomodar();
    }

    private void elegir(BuscarPc.Pc pc) {
        MiPc.guardar(this, pc.ip, pc.nombre);
        elegirAMano = false;
        aviso = null;
        web.loadUrl(MiPc.panel(pc.ip));
    }

    private void buscarPc() {
        mostrarAviso(BUSCANDO);
        if (buscando) {
            return;
        }
        buscando = true;
        new Thread(new Runnable() {
            @Override
            public void run() {
                BuscarPc.Resultado resultado = null;
                try {
                    resultado = BuscarPc.buscar(red);
                } catch (Exception e) {
                    Log.w("Menu", "no se pudo buscar la PC", e);
                }
                final BuscarPc.Resultado encontrado = resultado;
                runOnUiThread(new Runnable() {
                    @Override
                    public void run() {
                        buscando = false;
                        if (!isDestroyed()) {
                            terminoLaBusqueda(encontrado);
                        }
                    }
                });
            }
        }, "BuscarPc").start();
    }

    private void terminoLaBusqueda(BuscarPc.Resultado resultado) {
        hallado = resultado;
        if (!BUSCANDO.equals(aviso) || MiPc.ip(this) != null) {
            acomodar(); // mientras tanto cambió la red: se decide de nuevo
            return;
        }
        if (resultado == null || resultado.pcs.isEmpty()) {
            mostrarAviso(SIN_PC);
        } else if (resultado.pcs.size() == 1 && resultado.pcs.get(0).deMiCuenta && !elegirAMano) {
            elegir(resultado.pcs.get(0)); // la de siempre: sin preguntar
        } else {
            mostrarAviso(ELEGIR);
        }
    }

    @Override
    protected void onActivityResult(int pedido, int resultado, Intent datos) {
        super.onActivityResult(pedido, resultado, datos);
        if (pedido == PEDIDO_VPN) {
            red.conectar();
            acomodar();
        }
    }

    /** false si no hay nada que la abra (en la tele no suele haber navegador). */
    private boolean abrirAfuera(Uri url) {
        try {
            startActivity(new Intent(Intent.ACTION_VIEW, url));
            return true;
        } catch (ActivityNotFoundException e) {
            return false;
        }
    }

    static boolean esTele(Context contexto) {
        UiModeManager modo = (UiModeManager) contexto.getSystemService(Context.UI_MODE_SERVICE);
        return (modo != null && modo.getCurrentModeType() == Configuration.UI_MODE_TYPE_TELEVISION)
            || contexto.getPackageManager().hasSystemFeature(PackageManager.FEATURE_LEANBACK);
    }

    /** Lo que la página puede pedirle a la app. */
    private class ApiJs {
        @JavascriptInterface
        public void jugar(String uuid, String pc, String app) {
            final Uri url = new Uri.Builder().scheme("cloudgaming").authority("jugar")
                .appendQueryParameter("uuid", uuid)
                .appendQueryParameter("pc", pc)
                .appendQueryParameter("app", app)
                .build();
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    Lanzador.lanzar(Menu.this, url);
                }
            });
        }

        /** Ajustes del stream (resolución, bitrate, búfer de audio...) en JSON, para dibujarlos en el menú. */
        @JavascriptInterface
        public String ajustesStream() {
            try {
                return Ajustes.leer(Menu.this).toString();
            } catch (JSONException e) {
                return "null";
            }
        }

        @JavascriptInterface
        public boolean guardarAjuste(String clave, String valor) {
            return clave != null && valor != null && Ajustes.guardar(Menu.this, clave, valor);
        }

        /**
         * Cómo está la red de la app: «lista», «conectando», o lo que le falta. Para que el
         * panel distinga «todavía estoy conectando» de «la PC no contesta».
         */
        @JavascriptInterface
        public String red() {
            return red.estado();
        }

        @JavascriptInterface
        public String version() {
            return VERSION;
        }

        /** El APK de esta misma versión en el panel: el del celu (64 bits) o el de la tele (32). */
        @JavascriptInterface
        public String apk() {
            return BuildConfig.CLOUDGAMING_APK;
        }

        /** Baja e instala la actualización desde el panel (url relativa a él), sin navegador. */
        @JavascriptInterface
        public void actualizar(final String url) {
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    String panel = MiPc.panel(Menu.this);
                    if (panel != null) {
                        Actualizador.actualizar(Menu.this, Uri.parse(panel).buildUpon().encodedPath(url).build().toString());
                    }
                }
            });
        }

        /** «Cambiar de PC» del panel: vuelve a buscar las PC de la red para elegir una. */
        @JavascriptInterface
        public void cambiarPc() {
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    Menu.this.cambiarPc();
                }
            });
        }
    }

    // ------------------------------------------------------------------ joystick

    @Override
    public boolean dispatchKeyEvent(KeyEvent evento) {
        int tecla = traducir(evento.getKeyCode());
        if (tecla != 0 && esDeMando(evento)) {
            // Tecla nueva, sin origen "joystick": el panel la recibe como una tecla común.
            web.dispatchKeyEvent(new KeyEvent(evento.getDownTime(), evento.getEventTime(),
                evento.getAction(), tecla, evento.getRepeatCount()));
            return true;
        }
        return super.dispatchKeyEvent(evento);
    }

    @Override
    public boolean dispatchGenericMotionEvent(MotionEvent evento) {
        // Sin atender la palanca ni la cruceta analógica: Android las convierte solo en
        // teclas de dirección (con repetición), que llegan a dispatchKeyEvent.
        if ((evento.getSource() & InputDevice.SOURCE_JOYSTICK) == InputDevice.SOURCE_JOYSTICK) {
            return false;
        }
        return super.dispatchGenericMotionEvent(evento);
    }

    private static boolean esDeMando(KeyEvent evento) {
        int origen = evento.getSource();
        return (origen & InputDevice.SOURCE_GAMEPAD) == InputDevice.SOURCE_GAMEPAD
            || (origen & InputDevice.SOURCE_JOYSTICK) == InputDevice.SOURCE_JOYSTICK
            || (origen & InputDevice.SOURCE_DPAD) == InputDevice.SOURCE_DPAD;
    }

    private static int traducir(int codigo) {
        switch (codigo) {
            case KeyEvent.KEYCODE_DPAD_UP:
            case KeyEvent.KEYCODE_DPAD_DOWN:
            case KeyEvent.KEYCODE_DPAD_LEFT:
            case KeyEvent.KEYCODE_DPAD_RIGHT:
                return codigo;
            case KeyEvent.KEYCODE_BUTTON_L1:
                return KeyEvent.KEYCODE_DPAD_LEFT;
            case KeyEvent.KEYCODE_BUTTON_R1:
                return KeyEvent.KEYCODE_DPAD_RIGHT;
            case KeyEvent.KEYCODE_BUTTON_A:
            case KeyEvent.KEYCODE_BUTTON_START:
            case KeyEvent.KEYCODE_DPAD_CENTER:
                return KeyEvent.KEYCODE_ENTER;
            case KeyEvent.KEYCODE_BUTTON_B:
                return KeyEvent.KEYCODE_ESCAPE;
            default:
                return 0;
        }
    }

    @Override
    public void onBackPressed() {
        // Primero cierra la hoja abierta del panel; si no hay ninguna, manda la app al fondo.
        web.evaluateJavascript("window.panelAtras ? panelAtras() : false", new ValueCallback<String>() {
            @Override
            public void onReceiveValue(String resultado) {
                if (!"true".equals(resultado)) {
                    moveTaskToBack(true);
                }
            }
        });
    }

    // ------------------------------------------------------------- ciclo de vida

    /**
     * Dibujar también en el costado de la cámara. Si no, en horizontal Android deja ese
     * costado en negro (el stream de Moonlight hace lo mismo con «estirar la imagen»).
     */
    static void usarZonaCamara(Activity actividad) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.P) {
            return;
        }
        WindowManager.LayoutParams atributos = actividad.getWindow().getAttributes();
        atributos.layoutInDisplayCutoutMode = Build.VERSION.SDK_INT >= Build.VERSION_CODES.R
            ? WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_ALWAYS
            : WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES;
        actividad.getWindow().setAttributes(atributos);
    }

    @Override
    public void onWindowFocusChanged(boolean foco) {
        super.onWindowFocusChanged(foco);
        if (foco) {
            pantallaCompleta();
        }
    }

    @SuppressWarnings("deprecation")
    private void pantallaCompleta() {
        getWindow().getDecorView().setSystemUiVisibility(
            View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
                | View.SYSTEM_UI_FLAG_FULLSCREEN
                | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
                | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION);
    }

    @Override
    protected void onStart() {
        super.onStart();
        red.escuchar(alCambiarLaRed);
        if (SIN_PC.equals(aviso)) {
            aviso = null; // a lo mejor salió a instalar CloudGaming Host: se busca de nuevo
        }
        acomodar();
    }

    @Override
    protected void onStop() {
        red.dejarDeEscuchar(alCambiarLaRed);
        super.onStop();
    }

    @Override
    protected void onResume() {
        super.onResume();
        web.onResume();
        web.evaluateJavascript("window.panelEnPausa && panelEnPausa(false)", null);
    }

    @Override
    protected void onPause() {
        // Que la música del menú se calle (en segundo plano o durante el stream).
        web.evaluateJavascript("window.panelEnPausa && panelEnPausa(true)", null);
        web.onPause();
        CookieManager.getInstance().flush();
        super.onPause();
    }

    @Override
    protected void onSaveInstanceState(Bundle salida) {
        super.onSaveInstanceState(salida);
        if (aviso == null) { // una página propia no se guarda: al volver se decide de nuevo
            web.saveState(salida);
        }
    }

    @Override
    protected void onDestroy() {
        web.destroy();
        super.onDestroy();
    }
}
