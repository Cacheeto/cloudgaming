package ar.cloudgaming.app;

import android.app.Activity;
import android.content.Context;
import android.content.SharedPreferences;
import android.preference.PreferenceManager;
import android.view.Display;

import com.limelight.preferences.PreferenceConfiguration;

import org.json.JSONException;
import org.json.JSONObject;

/**
 * Ajustes del stream para el menú: son las mismas preferencias que lee Moonlight
 * (PreferenceConfiguration), pero se muestran y se cambian desde el panel.
 */
final class Ajustes {
    // Claves de Moonlight y sus valores por defecto (los de PreferenceConfiguration).
    private static final String RESOLUCION = "list_resolution";
    private static final String FPS = "list_fps";
    private static final String BITRATE = "seekbar_bitrate_kbps";
    private static final String BUFER_AUDIO = "list_audio_buffer";
    private static final String CODEC = "video_format";
    private static final String ESTIRAR = "checkbox_stretch_video";
    private static final String AUDIO_PC = "checkbox_host_audio";
    private static final String ESTADISTICAS = "checkbox_enable_perf_overlay";
    private static final String MODO_MOUSE = "checkbox_mouse_emulation";
    // Propia de CloudGaming (ver TecladoPc)
    private static final String TECLADO_PC = "cloudgaming_teclado_pc";

    private static final String RESOLUCION_DEFECTO = "1280x720";
    private static final String FPS_DEFECTO = "60";
    private static final String BUFER_DEFECTO = "80";
    static final String TECLADO_AUTOMATICO = "auto";

    private Ajustes() {
    }

    @SuppressWarnings("deprecation") // PreferenceManager: el mismo que usa Moonlight
    private static SharedPreferences prefs(Context contexto) {
        return PreferenceManager.getDefaultSharedPreferences(contexto);
    }

    static JSONObject leer(Activity actividad) throws JSONException {
        SharedPreferences p = prefs(actividad);
        String resolucion = p.getString(RESOLUCION, RESOLUCION_DEFECTO);
        String fps = p.getString(FPS, FPS_DEFECTO);
        Display.Mode modo = pantalla(actividad).getMode();
        float maxFps = 0;
        for (Display.Mode m : pantalla(actividad).getSupportedModes()) {
            maxFps = Math.max(maxFps, m.getRefreshRate());
        }
        int ancho = Math.max(modo.getPhysicalWidth(), modo.getPhysicalHeight());
        int alto = Math.min(modo.getPhysicalWidth(), modo.getPhysicalHeight());
        return new JSONObject()
            .put("resolucion", resolucion)
            .put("fps", fps)
            .put("bitrate", p.getInt(BITRATE, 0)) // 0 = automático
            .put("bitrateAuto", PreferenceConfiguration.getDefaultBitrate(resolucion, fps))
            .put("bufferAudio", Integer.parseInt(p.getString(BUFER_AUDIO, BUFER_DEFECTO)))
            .put("codec", p.getString(CODEC, "auto"))
            .put("estirar", p.getBoolean(ESTIRAR, false))
            .put("audioPc", p.getBoolean(AUDIO_PC, false))
            .put("estadisticas", p.getBoolean(ESTADISTICAS, false))
            .put("modoMouse", p.getBoolean(MODO_MOUSE, true))
            .put("tecladoPc", tecladoPc(actividad))
            .put("tecladoInformado", MiPc.teclado(actividad))
            .put("nativa", ancho + "x" + alto)
            .put("maxFps", Math.round(maxFps));
    }

    /** Devuelve false si la clave o el valor no son válidos (no guarda nada). */
    static boolean guardar(Context contexto, String clave, String valor) {
        SharedPreferences.Editor e = prefs(contexto).edit();
        switch (clave) {
            case "resolucion":
                if (!valor.matches("\\d{3,4}x\\d{3,4}")) return false;
                e.putString(RESOLUCION, valor);
                break;
            case "fps":
                if (!valor.matches("\\d{2,3}")) return false;
                e.putString(FPS, valor);
                break;
            case "bitrate":
                if ("auto".equals(valor)) {
                    e.remove(BITRATE); // Moonlight calcula el que corresponde a la resolución
                } else {
                    if (!valor.matches("\\d{3,6}")) return false;
                    e.putInt(BITRATE, Integer.parseInt(valor));
                }
                break;
            case "bufferAudio":
                if (!valor.matches("\\d{1,3}")) return false;
                e.putString(BUFER_AUDIO, valor);
                break;
            case "codec":
                if (!valor.matches("auto|forceh265|neverh265|forceav1")) return false;
                e.putString(CODEC, valor);
                break;
            case "estirar":
                e.putBoolean(ESTIRAR, Boolean.parseBoolean(valor));
                break;
            case "audioPc":
                e.putBoolean(AUDIO_PC, Boolean.parseBoolean(valor));
                break;
            case "estadisticas":
                e.putBoolean(ESTADISTICAS, Boolean.parseBoolean(valor));
                break;
            case "modoMouse":
                e.putBoolean(MODO_MOUSE, Boolean.parseBoolean(valor));
                break;
            case "tecladoPc":
                if (!TECLADO_AUTOMATICO.equals(valor) && !TecladoPc.conocida(valor)) return false;
                e.putString(TECLADO_PC, valor);
                break;
            default:
                return false;
        }
        e.apply();
        return true;
    }

    /** La distribución de teclado de la PC elegida a mano, o TECLADO_AUTOMATICO (la que informa la PC). */
    static String tecladoPc(Context contexto) {
        return prefs(contexto).getString(TECLADO_PC, TECLADO_AUTOMATICO);
    }

    @SuppressWarnings("deprecation") // getDefaultDisplay: minSdk 26, getDisplay() recién en 30
    private static Display pantalla(Activity actividad) {
        return actividad.getWindowManager().getDefaultDisplay();
    }
}
