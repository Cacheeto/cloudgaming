package ar.cloudgaming.app;

import android.content.Context;
import android.content.SharedPreferences;

/**
 * La PC con la que juega la app: la que se eligió en el asistente de la primera vez (ver
 * BuscarPc), o después con «Cambiar de PC». Se guarda su IP en la red (la de Tailscale).
 *
 * El panel se abre directo en esa IP y por http: la red ya cifra todo, y el panel reconoce
 * los equipos de la misma cuenta, así que no pide usuario ni contraseña.
 */
final class MiPc {
    static final int PUERTO_PANEL = 47777;

    private static final String ARCHIVO = "pc";
    private static final String IP = "ip";
    private static final String NOMBRE = "nombre";
    private static final String TECLADO = "teclado";

    private MiPc() {
    }

    /** La IP de la PC elegida, o null si todavía no hay ninguna. */
    static String ip(Context contexto) {
        return prefs(contexto).getString(IP, null);
    }

    /** La dirección del panel de la PC elegida, o null si todavía no hay ninguna. */
    static String panel(Context contexto) {
        String ip = ip(contexto);
        return ip != null ? panel(ip) : null;
    }

    static String panel(String ip) {
        return "http://" + ip + ":" + PUERTO_PANEL + "/";
    }

    static void guardar(Context contexto, String ip, String nombre) {
        prefs(contexto).edit().putString(IP, ip).putString(NOMBRE, nombre).apply();
    }

    /** La distribución de teclado que informó la PC al conectar («latam», «es»…), o "" (ver TecladoPc). */
    static String teclado(Context contexto) {
        return prefs(contexto).getString(TECLADO, "");
    }

    static void guardarTeclado(Context contexto, String teclado) {
        prefs(contexto).edit().putString(TECLADO, teclado).apply();
    }

    static void olvidar(Context contexto) {
        prefs(contexto).edit().clear().apply();
    }

    private static SharedPreferences prefs(Context contexto) {
        return contexto.getSharedPreferences(ARCHIVO, Context.MODE_PRIVATE);
    }
}
