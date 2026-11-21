package ar.cloudgaming.app;

import android.app.AlertDialog;
import android.content.Context;
import android.content.DialogInterface;
import android.content.Intent;
import android.widget.Toast;

import com.limelight.Game;
import com.limelight.binding.PlatformBinding;
import com.limelight.nvstream.http.ComputerDetails;
import com.limelight.nvstream.http.NvHTTP;

import org.xmlpull.v1.XmlPullParserException;

import java.io.ByteArrayInputStream;
import java.io.IOException;
import java.security.cert.CertificateException;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;

/**
 * Menú dentro del stream. Se abre con Atrás en el celu, o manteniendo un segundo
 * Guía (o Start+Select) en el mando. Se maneja con la cruceta, A elige y B cierra.
 */
public final class MenuJuego {
    private static final String[] OPCIONES = {
        "Seguir jugando",
        "Mostrar el teclado",
        "Volver al menú · el juego sigue abierto en la PC",
        "Terminar sesión · cierra el juego en la PC",
    };

    private static AlertDialog abierto;

    private MenuJuego() {
    }

    public static void mostrar(final Game juego) {
        if (juego.isFinishing() || (abierto != null && abierto.isShowing())) {
            return;
        }
        String app = juego.getIntent().getStringExtra(Game.EXTRA_APP_NAME);
        abierto = new AlertDialog.Builder(juego, android.R.style.Theme_DeviceDefault_Dialog_Alert)
            .setTitle(app != null ? app : "CloudGaming")
            .setItems(OPCIONES, new DialogInterface.OnClickListener() {
                @Override
                public void onClick(DialogInterface dialogo, int cual) {
                    switch (cual) {
                        case 1:
                            // Se muestra cuando el stream recupera el foco; desde acá
                            // Android lo esconde apenas se cierra el menú.
                            juego.showKeyboardAfterMenu();
                            break;
                        case 2:
                            juego.finish(); // corta el stream; la app sigue corriendo en Sunshine
                            break;
                        case 3:
                            terminar(juego);
                            break;
                        default:
                            break;
                    }
                }
            })
            .setOnDismissListener(new DialogInterface.OnDismissListener() {
                @Override
                public void onDismiss(DialogInterface dialogo) {
                    abierto = null;
                }
            })
            .show();
    }

    /** Le pide a Sunshine que cierre la app (como «Terminar sesión» del panel). */
    private static void terminar(final Game juego) {
        Intent datos = juego.getIntent();
        final String host = datos.getStringExtra(Game.EXTRA_HOST);
        final int puerto = datos.getIntExtra(Game.EXTRA_PORT, NvHTTP.DEFAULT_HTTP_PORT);
        final int puertoHttps = datos.getIntExtra(Game.EXTRA_HTTPS_PORT, 0);
        final String id = datos.getStringExtra(Game.EXTRA_UNIQUEID);
        final byte[] certificado = datos.getByteArrayExtra(Game.EXTRA_SERVER_CERT);
        final Context contexto = juego.getApplicationContext();
        Toast.makeText(juego, "Terminando la sesión…", Toast.LENGTH_SHORT).show();

        new Thread(new Runnable() {
            @Override
            public void run() {
                String error = null;
                try {
                    X509Certificate cert = certificado == null ? null : (X509Certificate)
                        CertificateFactory.getInstance("X.509").generateCertificate(new ByteArrayInputStream(certificado));
                    NvHTTP http = new NvHTTP(new ComputerDetails.AddressTuple(host, puerto), puertoHttps, id, cert,
                        PlatformBinding.getCryptoProvider(contexto));
                    if (!http.quitApp()) {
                        error = "la PC no la cerró";
                    }
                } catch (IOException | XmlPullParserException | CertificateException e) {
                    error = e.getMessage();
                }
                final String falla = error;
                juego.runOnUiThread(new Runnable() {
                    @Override
                    public void run() {
                        if (falla != null) {
                            Toast.makeText(contexto, "No se pudo terminar la sesión: " + falla, Toast.LENGTH_LONG).show();
                        } else {
                            // Sunshine ya cortó el stream; Game suele cerrarse solo, por las dudas lo cerramos.
                            juego.finish();
                        }
                    }
                });
            }
        }, "TerminarSesion").start();
    }
}
