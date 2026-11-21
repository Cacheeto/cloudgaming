package ar.cloudgaming.app;

import android.app.Activity;
import android.content.Intent;
import android.net.Uri;

import com.limelight.AppView;
import com.limelight.Game;

/**
 * Abre el stream dentro de esta misma app (el código de Moonlight viene incluido).
 *
 * Todo pasa por Preparar, que se asegura de que la PC esté guardada y emparejada y
 * busca el juego en la PC antes de arrancar.
 */
final class Lanzador {
    private Lanzador() {
    }

    /** cloudgaming://jugar?uuid=...&pc=...&app=... */
    static void lanzar(Activity actividad, Uri url) {
        if (!"jugar".equals(url.getAuthority())) {
            return;
        }
        Intent intento = new Intent(actividad, Preparar.class);
        poner(intento, AppView.UUID_EXTRA, url.getQueryParameter("uuid"));
        poner(intento, AppView.NAME_EXTRA, url.getQueryParameter("pc"));
        poner(intento, Game.EXTRA_APP_NAME, url.getQueryParameter("app"));
        actividad.startActivity(intento);
    }

    private static void poner(Intent intento, String clave, String valor) {
        if (valor != null && !valor.isEmpty()) {
            intento.putExtra(clave, valor);
        }
    }
}
