package ar.cloudgaming.app;

import android.app.Activity;
import android.net.Uri;
import android.os.Bundle;

/** Pantalla invisible: recibe cloudgaming://... desde el navegador, abre el stream y se cierra. */
public class Puente extends Activity {
    @Override
    protected void onCreate(Bundle guardado) {
        super.onCreate(guardado);
        Uri url = getIntent().getData();
        if (url != null) {
            Lanzador.lanzar(this, url);
        }
        finish();
    }
}
