package ar.cloudgaming.app;

import android.content.Context;
import android.view.KeyCharacterMap;
import android.view.KeyEvent;

import com.limelight.nvstream.NvConnection;
import com.limelight.nvstream.input.KeyboardPacket;

import java.text.Normalizer;
import java.util.HashMap;
import java.util.Map;

/**
 * Lo que se escribe con el teclado en pantalla del celu, escrito con la distribución de teclado
 * de la PC.
 *
 * Moonlight manda cada tecla por su lugar en un teclado de EE.UU., y la PC la interpreta con su
 * propia distribución: con la PC en latinoamericano, «@» (Shift+2 en EE.UU.) salía «"». Acá se
 * busca qué teclas hay que tocar en la distribución de la PC para cada carácter: con Shift, con
 * AltGr, y las tildes con su tecla muerta («á» = «´» y después «a»).
 *
 * La distribución la informa el panel de la PC (ver Preparar), o se elige en los ajustes del
 * stream. Si no se conoce, todo sigue como en Moonlight. Lo que la distribución no tiene va como
 * texto: Sunshine lo escribe con su código Unicode, que en Windows anda siempre y en Linux depende
 * del programa (en WoW no).
 *
 * Solo toca el teclado en pantalla: un teclado físico ya manda sus teclas por su lugar.
 */
public final class TecladoPc {
    // Teclas virtuales de Windows por su lugar en un teclado de EE.UU., que es lo que espera
    // Sunshine (las letras y los números usan su propio código ASCII en mayúscula).
    private static final int VK_ESPACIO = 0x20;
    private static final int VK_SHIFT = 0xA0; // el izquierdo
    private static final int VK_ALTGR = 0xA5; // Alt derecho
    private static final int OEM_1 = 0xBA;     // ; al lado de la L
    private static final int OEM_MAS = 0xBB;   // = al lado del 0
    private static final int OEM_COMA = 0xBC;
    private static final int OEM_MENOS = 0xBD;
    private static final int OEM_PUNTO = 0xBE;
    private static final int OEM_2 = 0xBF;     // / al lado del Shift derecho
    private static final int OEM_3 = 0xC0;     // ` arriba del Tab
    private static final int OEM_4 = 0xDB;     // [ al lado de la P
    private static final int OEM_5 = 0xDC;     // \ al lado del Enter
    private static final int OEM_6 = 0xDD;     // ] al lado del anterior
    private static final int OEM_7 = 0xDE;     // ' al lado de la Ñ
    private static final int OEM_102 = 0xE2;   // < al lado del Shift izquierdo

    // Una pulsación = la tecla (vk) más estos bits
    private static final int SHIFT = 0x100;
    private static final int ALTGR = 0x200;

    private static final char NADA = 0;

    /** El acento suelto de cada tilde combinable (la que pone la tecla muerta). */
    private static final Map<Character, Character> ACENTOS = new HashMap<>();
    /** Las distribuciones que se saben traducir, con el nombre que les da Linux (XKB). */
    private static final Map<String, Tabla> TABLAS = new HashMap<>();

    static {
        ACENTOS.put('́', '´');
        ACENTOS.put('̈', '¨');
        ACENTOS.put('̀', '`');
        ACENTOS.put('̂', '^');
        ACENTOS.put('̃', '~');

        TABLAS.put("latam", new Tabla()
            .numeros("!\"#$%&/()=")
            .tecla(OEM_3, '|', '°', '¬')
            .tecla(OEM_MENOS, '\'', '?', '\\')
            .tecla(OEM_MAS, '¿', '¡', NADA)
            .tecla('Q', 'q', 'Q', '@')
            .tecla(OEM_6, '+', '*', '~')
            .tecla(OEM_1, 'ñ', 'Ñ', NADA)
            .tecla(OEM_7, '{', '[', NADA)
            .tecla(OEM_5, '}', ']', NADA)
            .tecla(OEM_102, '<', '>', NADA)
            .tecla(OEM_COMA, ',', ';', NADA)
            .tecla(OEM_PUNTO, '.', ':', NADA)
            .tecla(OEM_2, '-', '_', NADA)
            .muerta(OEM_4, '´')
            .muerta(OEM_4 | SHIFT, '¨')
            .muerta(OEM_7 | ALTGR, '^')
            .muerta(OEM_5 | ALTGR, '`'));

        TABLAS.put("es", new Tabla()
            .numeros("!\"·$%&/()=")
            .tecla('1', '1', '!', '|')
            .tecla('2', '2', '"', '@')
            .tecla('3', '3', '·', '#')
            .tecla('4', '4', '$', '~')
            .tecla('6', '6', '&', '¬')
            .tecla(OEM_3, 'º', 'ª', '\\')
            .tecla(OEM_MENOS, '\'', '?', NADA)
            .tecla(OEM_MAS, '¡', '¿', NADA)
            .tecla('E', 'e', 'E', '€')
            .tecla(OEM_4, NADA, NADA, '[')
            .tecla(OEM_6, '+', '*', ']')
            .tecla(OEM_1, 'ñ', 'Ñ', NADA)
            .tecla(OEM_7, NADA, NADA, '{')
            .tecla(OEM_5, 'ç', 'Ç', '}')
            .tecla(OEM_102, '<', '>', NADA)
            .tecla(OEM_COMA, ',', ';', NADA)
            .tecla(OEM_PUNTO, '.', ':', NADA)
            .tecla(OEM_2, '-', '_', NADA)
            .muerta(OEM_4, '`')
            .muerta(OEM_4 | SHIFT, '^')
            .muerta(OEM_7, '´')
            .muerta(OEM_7 | SHIFT, '¨'));

        TABLAS.put("us", new Tabla()
            .numeros("!@#$%^&*()")
            .tecla(OEM_3, '`', '~', NADA)
            .tecla(OEM_MENOS, '-', '_', NADA)
            .tecla(OEM_MAS, '=', '+', NADA)
            .tecla(OEM_4, '[', '{', NADA)
            .tecla(OEM_6, ']', '}', NADA)
            .tecla(OEM_5, '\\', '|', NADA)
            .tecla(OEM_1, ';', ':', NADA)
            .tecla(OEM_7, '\'', '"', NADA)
            .tecla(OEM_COMA, ',', '<', NADA)
            .tecla(OEM_PUNTO, '.', '>', NADA)
            .tecla(OEM_2, '/', '?', NADA));
    }

    /** Qué escribe cada tecla en una distribución. */
    private static final class Tabla {
        private final Map<Integer, Integer> teclas = new HashMap<>();
        /** Acento suelto → la tecla muerta que lo pone. */
        private final Map<Integer, Integer> muertas = new HashMap<>();

        Tabla() {
            agregar(' ', VK_ESPACIO);
            for (char letra = 'a'; letra <= 'z'; letra++) {
                int vk = Character.toUpperCase(letra);
                agregar(letra, vk);
                agregar(Character.toUpperCase(letra), vk | SHIFT);
            }
            for (char numero = '0'; numero <= '9'; numero++) {
                agregar(numero, numero);
            }
        }

        /** Lo que escribe la fila de los números con Shift, del 1 al 0. */
        Tabla numeros(String conShift) {
            for (int i = 0; i < 10; i++) {
                agregar(conShift.charAt(i), (i < 9 ? '1' + i : '0') | SHIFT);
            }
            return this;
        }

        /** Lo que escribe una tecla sola, con Shift y con AltGr (NADA si no escribe nada). */
        Tabla tecla(int vk, char sola, char conShift, char conAltGr) {
            agregar(sola, vk);
            agregar(conShift, vk | SHIFT);
            agregar(conAltGr, vk | ALTGR);
            return this;
        }

        /** Una tecla muerta: pone el acento sobre la letra que sigue. */
        Tabla muerta(int pulsacion, char acento) {
            muertas.put((int) acento, pulsacion);
            return this;
        }

        /** Si un carácter sale de dos maneras, queda la más simple (sola, después Shift, después AltGr). */
        private void agregar(char caracter, int pulsacion) {
            Integer antes = teclas.get((int) caracter);
            if (caracter != NADA && (antes == null || nivel(pulsacion) < nivel(antes))) {
                teclas.put((int) caracter, pulsacion);
            }
        }

        private static int nivel(int pulsacion) {
            return (pulsacion & ALTGR) != 0 ? 2 : (pulsacion & SHIFT) != 0 ? 1 : 0;
        }

        /** Las pulsaciones que escriben el carácter, o null si con esta distribución no se puede. */
        int[] pulsaciones(int caracter) {
            Integer directa = teclas.get(caracter);
            if (directa != null) {
                return new int[] {directa};
            }
            // Letra con tilde: la tecla muerta del acento y después la letra
            String partes = Normalizer.normalize(new String(Character.toChars(caracter)), Normalizer.Form.NFD);
            if (partes.length() == 2) {
                Character acento = ACENTOS.get(partes.charAt(1));
                Integer muerta = acento != null ? muertas.get((int) acento) : null;
                Integer letra = teclas.get((int) partes.charAt(0));
                if (muerta != null && letra != null) {
                    return new int[] {muerta, letra};
                }
            }
            // El acento suelto: la tecla muerta y después espacio. Solo los que salen igual en
            // Windows y en Linux (ahí «´» y «¨» con espacio dan ' y ")
            Integer muerta = muertas.get(caracter);
            if (muerta != null && "^`~".indexOf(caracter) >= 0) {
                return new int[] {muerta, VK_ESPACIO};
            }
            return null;
        }
    }

    /** null: la distribución de la PC no se sabe traducir (todo sigue como en Moonlight). */
    private final Tabla tabla;
    /** Todo como texto: una PC con Windows y otra distribución (Windows escribe bien el texto). */
    private final boolean comoTexto;

    private TecladoPc(String distribucion) {
        tabla = TABLAS.get(distribucion);
        comoTexto = tabla == null && distribucion.startsWith("windows");
    }

    /** Con la distribución elegida en los ajustes, o si no la que informó la PC. */
    public static TecladoPc de(Context contexto) {
        String elegida = Ajustes.tecladoPc(contexto);
        return new TecladoPc(Ajustes.TECLADO_AUTOMATICO.equals(elegida) ? MiPc.teclado(contexto) : elegida);
    }

    static boolean conocida(String distribucion) {
        return TABLAS.containsKey(distribucion);
    }

    private boolean activo() {
        return tabla != null || comoTexto;
    }

    /**
     * Para Game, con cada tecla que llega. Si viene del teclado en pantalla y escribe un
     * carácter, lo manda a la PC con las teclas de su distribución y devuelve true. Con false,
     * Moonlight la maneja como siempre.
     */
    public boolean escribir(KeyEvent evento, NvConnection conexion) {
        if (!activo() || !delTecladoEnPantalla(evento)) {
            return false;
        }
        if (KeyEvent.isModifierKey(evento.getKeyCode())) {
            // El Shift que manda el teclado del celu es el de EE.UU.: el que haga falta lo toca
            // escribir(). Las teclas que no escriben (flechas, borrar) llevan el modificador en
            // el evento y Sunshine lo agrega solo.
            return true;
        }
        if (evento.isCtrlPressed() || evento.isAltPressed() || evento.isMetaPressed()) {
            return false; // un atajo (Ctrl+C): va por su lugar, como siempre
        }
        int caracter = evento.getUnicodeChar();
        if (caracter < ' ' || (caracter & KeyCharacterMap.COMBINING_ACCENT) != 0) {
            return false; // no escribe nada (Enter, Tab, borrar, flechas)
        }
        if (evento.getAction() == KeyEvent.ACTION_DOWN) {
            escribir(caracter, conexion);
        }
        return true; // al soltar no hay nada que mandar: escribir() ya tocó y soltó todo
    }

    /** Para Game: los caracteres que llegan juntos, sin tecla (los que Android no tiene en su teclado). */
    public boolean escribirTexto(String texto, NvConnection conexion) {
        if (!activo()) {
            return false;
        }
        for (int i = 0; i < texto.length(); i += Character.charCount(texto.codePointAt(i))) {
            escribir(texto.codePointAt(i), conexion);
        }
        return true;
    }

    private void escribir(int caracter, NvConnection conexion) {
        int[] pulsaciones = tabla != null ? tabla.pulsaciones(caracter) : null;
        if (pulsaciones == null) {
            conexion.sendUtf8Text(new String(Character.toChars(caracter)));
            return;
        }
        for (int pulsacion : pulsaciones) {
            boolean shift = (pulsacion & SHIFT) != 0;
            boolean altGr = (pulsacion & ALTGR) != 0;
            if (shift) tocar(conexion, VK_SHIFT, KeyboardPacket.KEY_DOWN);
            if (altGr) tocar(conexion, VK_ALTGR, KeyboardPacket.KEY_DOWN);
            tocar(conexion, pulsacion & 0xFF, KeyboardPacket.KEY_DOWN);
            tocar(conexion, pulsacion & 0xFF, KeyboardPacket.KEY_UP);
            if (altGr) tocar(conexion, VK_ALTGR, KeyboardPacket.KEY_UP);
            if (shift) tocar(conexion, VK_SHIFT, KeyboardPacket.KEY_UP);
        }
    }

    /**
     * Sin modificadores en el paquete: si llevara Alt, Sunshine tocaría además el Alt izquierdo,
     * que no es AltGr. Va como tecla «normalizada» (por su lugar en EE.UU.), así Sunshine en
     * Windows la pasa a la tecla de ese lugar y Windows la lee con su distribución.
     */
    private static void tocar(NvConnection conexion, int vk, byte direccion) {
        conexion.sendKeyboardInput((short) (0x8000 | vk), direccion, (byte) 0, (byte) 0);
    }

    private static boolean delTecladoEnPantalla(KeyEvent evento) {
        return evento.getDeviceId() == KeyCharacterMap.VIRTUAL_KEYBOARD
            || (evento.getFlags() & KeyEvent.FLAG_SOFT_KEYBOARD) != 0;
    }
}
