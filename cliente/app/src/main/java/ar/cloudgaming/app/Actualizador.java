package ar.cloudgaming.app;

import android.app.Activity;
import android.content.ActivityNotFoundException;
import android.content.ContentProvider;
import android.content.ContentValues;
import android.content.Context;
import android.content.Intent;
import android.content.pm.PackageInfo;
import android.database.Cursor;
import android.database.MatrixCursor;
import android.net.Uri;
import android.os.ParcelFileDescriptor;
import android.provider.OpenableColumns;
import android.provider.Settings;
import android.widget.Toast;

import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;

/**
 * Actualiza la app desde el panel sin pasar por el navegador. Desde la 3.0 la red a la PC
 * es solo de esta app: el navegador del celu no llega, y la tele ni siquiera tiene uno.
 * Baja el APK por esa red y se lo da al instalador de Android, que pide confirmar.
 *
 * También es el proveedor que le presta el APK bajado al instalador (content://), que no
 * puede leer los archivos de la app. Se usa el instalador común y no PackageInstaller
 * porque en MIUI/HyperOS este último falla con las optimizaciones de MIUI prendidas.
 */
public final class Actualizador extends ContentProvider {
    private static final String TIPO_APK = "application/vnd.android.package-archive";
    private static final String ARCHIVO = "actualizacion.apk";

    private static volatile boolean bajando;

    static boolean esApk(String url, String tipo) {
        return TIPO_APK.equals(tipo) || Uri.parse(url).getPath().endsWith(".apk");
    }

    static void actualizar(final Activity actividad, final String url) {
        if (bajando) {
            Toast.makeText(actividad, "La actualización ya se está bajando…", Toast.LENGTH_SHORT).show();
            return;
        }
        if (!actividad.getPackageManager().canRequestPackageInstalls()) {
            Toast.makeText(actividad, "Permití que CloudGaming instale apps y volvé a tocar «Descargar»",
                Toast.LENGTH_LONG).show();
            pedirPermiso(actividad);
            return;
        }
        bajando = true;
        Toast.makeText(actividad, "Bajando la actualización…", Toast.LENGTH_LONG).show();
        final Context app = actividad.getApplicationContext();
        new Thread(new Runnable() {
            @Override
            public void run() {
                String error = null;
                File apk = new File(app.getCacheDir(), ARCHIVO);
                try {
                    bajar(url, apk);
                    PackageInfo info = app.getPackageManager().getPackageArchiveInfo(apk.getPath(), 0);
                    if (info == null || !app.getPackageName().equals(info.packageName)) {
                        error = "lo que llegó no es la app";
                    }
                } catch (IOException e) {
                    error = e.getMessage();
                } finally {
                    bajando = false;
                }
                final String falla = error;
                actividad.runOnUiThread(new Runnable() {
                    @Override
                    public void run() {
                        if (falla != null) {
                            Toast.makeText(app, "No se pudo bajar la actualización: " + falla, Toast.LENGTH_LONG).show();
                        } else {
                            instalar(actividad);
                        }
                    }
                });
            }
        }, "Actualizador").start();
    }

    private static void bajar(String url, File destino) throws IOException {
        HttpURLConnection conexion = (HttpURLConnection) new URL(url).openConnection();
        conexion.setConnectTimeout(15000);
        conexion.setReadTimeout(30000);
        try {
            int codigo = conexion.getResponseCode();
            if (codigo != HttpURLConnection.HTTP_OK) {
                throw new IOException("la PC respondió " + codigo);
            }
            try (InputStream entrada = conexion.getInputStream();
                 OutputStream salida = new FileOutputStream(destino)) {
                byte[] bloque = new byte[64 * 1024];
                int leidos;
                while ((leidos = entrada.read(bloque)) != -1) {
                    salida.write(bloque, 0, leidos);
                }
            }
        } finally {
            conexion.disconnect();
        }
    }

    private static void instalar(Activity actividad) {
        Uri apk = new Uri.Builder().scheme("content").authority(autoridad(actividad)).path(ARCHIVO).build();
        Intent instalar = new Intent(Intent.ACTION_VIEW)
            .setDataAndType(apk, TIPO_APK)
            .addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        try {
            actividad.startActivity(instalar);
        } catch (ActivityNotFoundException e) {
            Toast.makeText(actividad, "Este equipo no tiene instalador de apps", Toast.LENGTH_LONG).show();
        }
    }

    /** «Instalar apps desconocidas» para esta app; si el equipo no tiene esa pantalla, la de seguridad. */
    private static void pedirPermiso(Activity actividad) {
        try {
            actividad.startActivity(new Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES,
                Uri.parse("package:" + actividad.getPackageName())));
        } catch (ActivityNotFoundException e) {
            try {
                actividad.startActivity(new Intent(Settings.ACTION_SECURITY_SETTINGS));
            } catch (ActivityNotFoundException ignorada) {
                // sin pantalla para darlo: queda el aviso
            }
        }
    }

    private static String autoridad(Context contexto) {
        return contexto.getPackageName() + ".actualizacion";
    }

    // ------------------------------------------------- proveedor para el instalador

    @Override
    public boolean onCreate() {
        return true;
    }

    @Override
    public String getType(Uri uri) {
        return TIPO_APK;
    }

    @Override
    public ParcelFileDescriptor openFile(Uri uri, String modo) throws FileNotFoundException {
        if (!ARCHIVO.equals(uri.getLastPathSegment()) || !"r".equals(modo)) {
            throw new FileNotFoundException(uri.toString());
        }
        return ParcelFileDescriptor.open(new File(getContext().getCacheDir(), ARCHIVO),
            ParcelFileDescriptor.MODE_READ_ONLY);
    }

    /** Nombre y tamaño, que algunos instaladores preguntan antes de abrirlo. */
    @Override
    public Cursor query(Uri uri, String[] columnas, String seleccion, String[] argumentos, String orden) {
        MatrixCursor fila = new MatrixCursor(new String[] {OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE});
        fila.addRow(new Object[] {"cloudgaming.apk", new File(getContext().getCacheDir(), ARCHIVO).length()});
        return fila;
    }

    @Override
    public Uri insert(Uri uri, ContentValues valores) {
        return null;
    }

    @Override
    public int delete(Uri uri, String seleccion, String[] argumentos) {
        return 0;
    }

    @Override
    public int update(Uri uri, ContentValues valores, String seleccion, String[] argumentos) {
        return 0;
    }
}
