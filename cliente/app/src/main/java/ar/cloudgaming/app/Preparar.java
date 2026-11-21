package ar.cloudgaming.app;

import android.app.Activity;
import android.content.ComponentName;
import android.content.Intent;
import android.content.ServiceConnection;
import android.graphics.Color;
import android.os.Bundle;
import android.os.IBinder;
import android.view.Gravity;
import android.view.View;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.ProgressBar;
import android.widget.TextView;

import com.limelight.AppView;
import com.limelight.Game;
import com.limelight.ShortcutTrampoline;
import com.limelight.binding.PlatformBinding;
import com.limelight.computers.ComputerManagerService;
import com.limelight.nvstream.http.ComputerDetails;
import com.limelight.nvstream.http.NvApp;
import com.limelight.nvstream.http.NvHTTP;
import com.limelight.nvstream.http.PairingManager;
import com.limelight.utils.CacheHelper;

import org.json.JSONObject;
import org.xmlpull.v1.XmlPullParserException;

import java.io.IOException;
import java.io.OutputStream;
import java.io.StringReader;

/**
 * Paso previo a cada stream:
 *  0. espera a que la red de la app esté lista (ver Red);
 *  1. si la PC no está guardada (o nunca se emparejó desde esta app), la agrega por su IP en esa red
 *     (la de la PC elegida en el menú, ver MiPc);
 *  2. si no está emparejada, la empareja pasándole el PIN al panel solo;
 *  3. pide la lista de juegos a la PC (y la deja en caché, como hace Moonlight) para sacar el
 *     número del juego, que es lo que necesita el stream.
 * Después sigue a ShortcutTrampoline, que abre el stream (o avisa si hay otra app abierta).
 */
public class Preparar extends Activity {
    private static final long ESPERA_RED_MS = 20 * 1000;

    private TextView texto;
    private ProgressBar rueda;
    private Button volver;
    private Red red;
    private boolean enlazado;

    private final ServiceConnection conexion = new ServiceConnection() {
        @Override
        public void onServiceConnected(ComponentName nombre, IBinder binder) {
            final ComputerManagerService.ComputerManagerBinder servicio =
                (ComputerManagerService.ComputerManagerBinder) binder;
            new Thread(new Runnable() {
                @Override
                public void run() {
                    preparar(servicio);
                }
            }, "Preparar").start();
        }

        @Override
        public void onServiceDisconnected(ComponentName nombre) {
        }
    };

    @Override
    protected void onCreate(Bundle guardado) {
        super.onCreate(guardado);
        Menu.usarZonaCamara(this);
        red = Red.de(this); // acá y no más tarde: si se llegó sin pasar por el menú, la arranca
        LinearLayout caja = new LinearLayout(this);
        caja.setOrientation(LinearLayout.VERTICAL);
        caja.setGravity(Gravity.CENTER);
        caja.setPadding(48, 48, 48, 48);
        caja.setBackgroundColor(Color.rgb(4, 10, 28));
        rueda = new ProgressBar(this);
        texto = new TextView(this);
        texto.setTextColor(Color.rgb(243, 246, 255));
        texto.setTextSize(18);
        texto.setGravity(Gravity.CENTER);
        texto.setPadding(0, 32, 0, 32);
        texto.setText("Conectando con la PC…");
        volver = new Button(this);
        volver.setText("Volver");
        volver.setVisibility(View.GONE);
        volver.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View vista) {
                finish();
            }
        });
        caja.addView(rueda);
        caja.addView(texto);
        caja.addView(volver);
        setContentView(caja);

        enlazado = bindService(new Intent(this, ComputerManagerService.class), conexion, BIND_AUTO_CREATE);
    }

    private void preparar(ComputerManagerService.ComputerManagerBinder servicio) {
        String juego = getIntent().getStringExtra(Game.EXTRA_APP_NAME);
        try {
            if (!red.esperarLista(ESPERA_RED_MS)) {
                fallar(Red.CONECTANDO.equals(red.estado())
                    ? "No pude conectar con tu red.\nRevisá que el celu tenga internet."
                    : "Falta terminar de configurar la conexión.\nAbrí CloudGaming y seguí los pasos.");
                return;
            }
            String ip = MiPc.ip(this);
            if (ip == null) {
                fallar("Todavía no elegiste tu PC.\nAbrí CloudGaming y seguí los pasos.");
                return;
            }
            // Qué distribución de teclado usa la PC, para que el teclado del celu escriba bien (ver TecladoPc)
            JSONObject hola = BuscarPc.hola(ip);
            if (hola != null) {
                MiPc.guardarTeclado(this, hola.optString("teclado", ""));
            }
            servicio.waitForReady();

            String uuid = getIntent().getStringExtra(AppView.UUID_EXTRA);
            ComputerDetails pc = uuid != null ? servicio.getComputer(uuid) : null;
            if (pc == null || pc.serverCert == null) {
                mostrar("Buscando la PC…");
                pc = new ComputerDetails();
                pc.manualAddress = new ComputerDetails.AddressTuple(ip, NvHTTP.DEFAULT_HTTP_PORT);
                if (!servicio.addComputerBlocking(pc)) {
                    fallar("No llego a la PC (" + ip + ").\nRevisá que esté prendida y con internet.");
                    return;
                }
            }

            // Siempre por la IP de la red propia: anda igual en casa y afuera.
            NvHTTP http = new NvHTTP(new ComputerDetails.AddressTuple(ip, NvHTTP.DEFAULT_HTTP_PORT),
                pc.httpsPort, servicio.getUniqueId(), pc.serverCert, PlatformBinding.getCryptoProvider(this));
            if (http.getPairState() != PairingManager.PairState.PAIRED) {
                String pin = PairingManager.generatePinString();
                mostrar("Emparejando este celu con la PC…\n\nPIN " + pin
                    + "\n\nSe lo paso solo al panel. Si en unos segundos no avanza,\ncargalo en el panel (Dispositivos → Emparejar).");
                PinAutomatico.enviar(this, pin);
                PairingManager emparejador = http.getPairingManager();
                PairingManager.PairState estado = emparejador.pair(http.getServerInfo(true), pin);
                if (estado != PairingManager.PairState.PAIRED) {
                    fallar("No se pudo emparejar (" + estado + ").");
                    return;
                }
                // Lo mismo que hace PcView al emparejar: fijar el certificado y volver a consultar.
                // (pair() ya se lo dejó puesto a http para lo que sigue.)
                ComputerDetails guardada = servicio.getComputer(pc.uuid);
                if (guardada != null) {
                    guardada.serverCert = emparejador.getPairedCert();
                }
                servicio.invalidateStateForComputer(pc.uuid);
            }

            mostrar("Abriendo " + (juego != null ? juego : "la PC") + "…");
            String lista = http.getAppListRaw();
            try (OutputStream cache = CacheHelper.openCacheFileForOutput(getCacheDir(), "applist", pc.uuid)) {
                CacheHelper.writeStringToOutputStream(cache, lista);
            }
            Intent seguir = new Intent(this, ShortcutTrampoline.class)
                .putExtra(AppView.UUID_EXTRA, pc.uuid)
                .putExtra(AppView.NAME_EXTRA, pc.name);
            if (juego != null) {
                NvApp app = buscar(lista, juego);
                if (app == null) {
                    fallar("No encuentro «" + juego + "» en la PC.");
                    return;
                }
                seguir.putExtra(Game.EXTRA_APP_NAME, app.getAppName())
                    .putExtra(Game.EXTRA_APP_ID, String.valueOf(app.getAppId()))
                    .putExtra(Game.EXTRA_APP_HDR, app.isHdrSupported());
            }
            startActivity(seguir);
            finish();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        } catch (IOException | XmlPullParserException e) {
            fallar("Error al conectar: " + e.getMessage());
        }
    }

    private static NvApp buscar(String lista, String nombre) throws IOException, XmlPullParserException {
        for (NvApp app : NvHTTP.getAppListByReader(new StringReader(lista))) {
            if (app.getAppName().equals(nombre)) {
                return app;
            }
        }
        return null;
    }

    private void mostrar(final String mensaje) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                texto.setText(mensaje);
            }
        });
    }

    private void fallar(final String mensaje) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                rueda.setVisibility(View.GONE);
                texto.setText(mensaje);
                volver.setVisibility(View.VISIBLE);
                volver.requestFocus();
            }
        });
    }

    @Override
    protected void onDestroy() {
        if (enlazado) {
            unbindService(conexion);
        }
        super.onDestroy();
    }
}
