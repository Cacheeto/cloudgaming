package ar.cloudgaming.app;

import android.app.Activity;
import android.webkit.CookieManager;

import com.limelight.LimeLog;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;

/**
 * Le pasa solo el PIN de emparejamiento al panel de la PC (que se lo da a Sunshine),
 * así no hay que tipearlo. El panel reconoce los equipos de la misma cuenta de la red;
 * si el celu entró con usuario y contraseña, va también la sesión del WebView del menú.
 *
 * Si no puede (no hay PC elegida, el panel no lo deja, o hay más de un pedido y no sé
 * cuál es el mío) no hace nada: el PIN sigue en pantalla para cargarlo a mano.
 */
public final class PinAutomatico {
    /** Uno por segundo: Sunshine anota el pedido recién cuando le llega el primer paso. */
    private static final int INTENTOS = 30;
    private static final String NOMBRE = "App CloudGaming";

    private PinAutomatico() {
    }

    public static void enviar(final Activity actividad, final String pin) {
        // CookieManager se usa desde el hilo principal; la red, en un hilo aparte.
        actividad.runOnUiThread(new Runnable() {
            @Override
            public void run() {
                final String panel = MiPc.panel(actividad);
                if (panel == null) {
                    LimeLog.warning("PIN automático: no hay PC elegida");
                    return;
                }
                final String galleta = CookieManager.getInstance().getCookie(panel);
                new Thread(new Runnable() {
                    @Override
                    public void run() {
                        probar(panel, galleta, pin);
                    }
                }, "PinAutomatico").start();
            }
        });
    }

    private static void probar(String panel, String galleta, String pin) {
        try {
            for (int i = 0; i < INTENTOS; i++) {
                Thread.sleep(1000);
                JSONArray pendientes = new JSONObject(pedir(panel, "GET", "api/pin", galleta, null))
                    .getJSONArray("pendientes");
                if (pendientes.length() == 1) {
                    JSONObject cuerpo = new JSONObject()
                        .put("pairing_id", String.valueOf(pendientes.getJSONObject(0).get("id")))
                        .put("pin", pin)
                        .put("name", NOMBRE);
                    pedir(panel, "POST", "api/pin", galleta, cuerpo.toString());
                    LimeLog.info("PIN automático: enviado al panel");
                    return;
                }
                if (pendientes.length() > 1) {
                    LimeLog.warning("PIN automático: hay varios pedidos, va a mano");
                    return;
                }
            }
        } catch (IOException | JSONException e) {
            LimeLog.warning("PIN automático: " + e);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }

    private static String pedir(String panel, String metodo, String ruta, String galleta, String cuerpo)
            throws IOException {
        HttpURLConnection conexion = (HttpURLConnection) new URL(panel + ruta).openConnection();
        try {
            conexion.setRequestMethod(metodo);
            conexion.setConnectTimeout(5000);
            conexion.setReadTimeout(10000);
            if (galleta != null) {
                conexion.setRequestProperty("Cookie", galleta);
            }
            if (cuerpo != null) {
                // X-Panel: el mismo encabezado anti-CSRF que manda la página.
                conexion.setDoOutput(true);
                conexion.setRequestProperty("Content-Type", "application/json");
                conexion.setRequestProperty("X-Panel", "1");
                try (OutputStream salida = conexion.getOutputStream()) {
                    salida.write(cuerpo.getBytes(StandardCharsets.UTF_8));
                }
            }
            if (conexion.getResponseCode() != HttpURLConnection.HTTP_OK) {
                throw new IOException("el panel respondió " + conexion.getResponseCode());
            }
            try (InputStream entrada = conexion.getInputStream()) {
                ByteArrayOutputStream datos = new ByteArrayOutputStream();
                byte[] bloque = new byte[4096];
                int leidos;
                while ((leidos = entrada.read(bloque)) != -1) {
                    datos.write(bloque, 0, leidos);
                }
                return datos.toString("UTF-8");
            }
        } finally {
            conexion.disconnect();
        }
    }
}
