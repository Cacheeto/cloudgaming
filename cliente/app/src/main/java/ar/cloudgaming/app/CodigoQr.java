package ar.cloudgaming.app;

import io.nayuki.qrcodegen.QrCode;

/**
 * Código QR como SVG, para las páginas propias de Menu. En la tele no hay navegador:
 * la página de inicio de sesión de la red se abre escaneándolo con el celu.
 */
final class CodigoQr {
    /** Margen blanco alrededor, en módulos (el estándar pide 4 para que lo lean bien). */
    private static final int BORDE = 4;

    private CodigoQr() {
    }

    static String svg(String texto) {
        QrCode qr = QrCode.encodeText(texto, QrCode.Ecc.MEDIUM);
        int lado = qr.size + BORDE * 2;
        StringBuilder trazo = new StringBuilder();
        for (int y = 0; y < qr.size; y++) {
            for (int x = 0; x < qr.size; x++) {
                if (qr.getModule(x, y)) {
                    trazo.append('M').append(x + BORDE).append(',').append(y + BORDE).append("h1v1h-1z");
                }
            }
        }
        return "<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 " + lado + ' ' + lado
            + "' shape-rendering='crispEdges'><rect width='100%' height='100%' fill='#fff'/>"
            + "<path d='" + trazo + "' fill='#000'/></svg>";
    }
}
