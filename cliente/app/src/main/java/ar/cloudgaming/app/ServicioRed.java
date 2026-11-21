package ar.cloudgaming.app;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.net.IpPrefix;
import android.net.Network;
import android.net.VpnService;
import android.os.Build;
import android.system.OsConstants;
import android.util.Log;

import com.limelight.R;

import java.net.InetAddress;
import java.util.UUID;

/**
 * El túnel hacia la PC. Para Android es una VPN, pero solo la usa esta app: el resto
 * del celu sigue saliendo a internet como siempre.
 *
 * Android entrega acá los paquetes de la app y el motor de red (ver Red) los cifra y
 * los manda. Es el IPNService de tailscale-android, reducido a lo que hace falta.
 */
public class ServicioRed extends VpnService implements libtailscale.IPNService {
    /** Armar el túnel. */
    static final String CONECTAR = "ar.cloudgaming.app.red.CONECTAR";
    /**
     * Solo quedar a la vista de Android, sin túnel: mientras se inicia sesión en el
     * navegador, para que no congele la app y el motor pueda enterarse de que terminó.
     */
    static final String SOLO_AVISO = "ar.cloudgaming.app.red.SOLO_AVISO";

    private static final String TAG = "ServicioRed";
    private static final String CANAL = "red";
    private static final int AVISO = 1;

    // Uno distinto por cada vez que se crea el servicio: así el motor sabe si es el mismo.
    private final String id = UUID.randomUUID().toString();
    private Red red;
    private boolean cerrado;

    @Override
    public void onCreate() {
        super.onCreate();
        red = Red.de(this);
        red.servicioCreado(this);
    }

    @Override
    public int onStartCommand(Intent intento, int banderas, int idInicio) {
        avisar();
        if (intento == null || !SOLO_AVISO.equals(intento.getAction())) {
            // CONECTAR, o Android mismo si se eligió «VPN siempre activa» para esta app.
            red.pedirTunel(this);
        }
        return START_NOT_STICKY;
    }

    /** Android exige un aviso fijo mientras el servicio trabaje con la app en segundo plano. */
    private void avisar() {
        getSystemService(NotificationManager.class).createNotificationChannel(
            new NotificationChannel(CANAL, "Conexión con la PC", NotificationManager.IMPORTANCE_LOW));
        Notification aviso = new Notification.Builder(this, CANAL)
            .setSmallIcon(R.drawable.ic_computer)
            .setContentTitle("CloudGaming")
            .setContentText("Conectado con tu PC")
            .setContentIntent(abrirMenu())
            .setOngoing(true)
            .build();
        try {
            startForeground(AVISO, aviso);
        } catch (RuntimeException e) {
            // Sin el permiso de VPN todavía, Android no lo acepta. No es grave: sigue
            // como servicio común.
            Log.w(TAG, "sin aviso fijo: " + e);
        }
    }

    private PendingIntent abrirMenu() {
        return PendingIntent.getActivity(this, 0, new Intent(this, Menu.class),
            PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE);
    }

    /** Cambió la red por la que sale el túnel (wifi, datos): avisarle a Android. */
    void salirPor(Network red) {
        setUnderlyingNetworks(red != null ? new Network[] {red} : new Network[0]);
    }

    @Override
    public void onRevoke() {
        // Otra app de VPN tomó el lugar (Android permite una sola a la vez).
        red.apagarMotor();
        close();
        super.onRevoke();
    }

    @Override
    public void onDestroy() {
        close();
        super.onDestroy();
    }

    // ------------------------------------------- lo que le pide el motor de red

    @Override
    public String id() {
        return id;
    }

    // protect(int), que deja un socket del motor fuera del túnel, ya lo trae VpnService.

    @Override
    public libtailscale.VPNServiceBuilder newBuilder() {
        Builder b = new Builder()
            .setSession("CloudGaming")
            .setConfigureIntent(abrirMenu())
            .allowFamily(OsConstants.AF_INET)
            .allowFamily(OsConstants.AF_INET6);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            b.setMetered(false); // que cuente como medida solo si lo es la red de abajo
        }
        Network salida = red.redDeSalida();
        b.setUnderlyingNetworks(salida != null ? new Network[] {salida} : new Network[0]);
        try {
            b.addAllowedApplication(getPackageName()); // el túnel es solo para esta app
        } catch (PackageManager.NameNotFoundException e) {
            throw new IllegalStateException(e);
        }
        return new Constructor(b);
    }

    @Override
    public void updateVpnStatus(boolean armado) {
        red.tunelArmado(armado);
    }

    @Override
    public void disconnectVPN() {
        stopSelf();
    }

    /** Baja el túnel y termina el servicio. */
    @Override
    public void close() {
        if (cerrado) {
            return;
        }
        cerrado = true;
        red.servicioCerrado(this);
        stopForeground(STOP_FOREGROUND_REMOVE);
        stopSelf();
    }

    /** Arma la interfaz del túnel con las direcciones y rutas que le va pasando el motor. */
    private static final class Constructor implements libtailscale.VPNServiceBuilder {
        private final Builder b;

        Constructor(Builder b) {
            this.b = b;
        }

        @Override
        public void setMTU(int mtu) {
            b.setMtu(mtu);
        }

        @Override
        public void addAddress(String ip, int bits) {
            b.addAddress(ip, bits);
        }

        @Override
        public void addRoute(String ip, int bits) {
            b.addRoute(ip, bits);
        }

        @Override
        public void excludeRoute(String ip, int bits) throws Exception {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                b.excludeRoute(new IpPrefix(InetAddress.getByName(ip), bits));
            }
        }

        @Override
        public void addDNSServer(String ip) {
            b.addDnsServer(ip);
        }

        @Override
        public void addSearchDomain(String dominio) {
            b.addSearchDomain(dominio);
        }

        @Override
        public libtailscale.ParcelFileDescriptor establish() {
            final android.os.ParcelFileDescriptor tunel = b.establish();
            if (tunel == null) {
                return null; // falta el permiso de VPN, o lo revocaron
            }
            return new libtailscale.ParcelFileDescriptor() {
                @Override
                public int detach() {
                    return tunel.detachFd();
                }
            };
        }
    }
}
