package ar.cloudgaming.app;

import android.net.ConnectivityManager;
import android.net.LinkProperties;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.NetworkRequest;
import android.net.RouteInfo;
import android.os.ParcelFileDescriptor;

import java.net.Inet4Address;
import java.net.InetAddress;
import java.util.LinkedHashMap;
import java.util.Map;

import libtailscale.Libtailscale;

/**
 * Sigue las redes del celu (wifi, datos) y le cuenta al motor de red cuál es la que sale
 * a internet: por esa manda sus paquetes y de ahí toma los DNS.
 *
 * Es el NetworkChangeCallback de tailscale-android (1.104) pasado a Java. Dos mejoras
 * tomadas de su versión siguiente: no se publica una red recién elegida hasta conocer
 * sus datos, y se avisa también cuando cambia la interfaz aunque los DNS sean los mismos.
 */
final class VigiaRed extends ConnectivityManager.NetworkCallback {
    interface Oyente {
        /** Cambió la red por la que se sale a internet (null: no hay ninguna). */
        void cambioLaRed(Network red);
    }

    private static final class Datos {
        NetworkCapabilities capacidades;
        LinkProperties propiedades;
    }

    private final Oyente oyente;
    // Redes con internet que no son una VPN, en el orden en que aparecieron.
    private final Map<Network, Datos> activas = new LinkedHashMap<>();
    private volatile Network actual;
    // Lo último que se le contó al motor, y los DNS de esa vez (ver dns()).
    private String publicado;
    private String dns = "";

    VigiaRed(Oyente oyente) {
        this.oyente = oyente;
    }

    void empezar(ConnectivityManager conectividad) {
        // No sirve registerDefaultNetworkCallback: con el túnel armado, la red por defecto
        // de esta app es el túnel mismo, y usarla para salir sería un círculo.
        conectividad.registerNetworkCallback(new NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .addCapability(NetworkCapabilities.NET_CAPABILITY_NOT_VPN)
            .build(), this);
    }

    /** La red por la que se sale a internet, o null. */
    Network actual() {
        return actual;
    }

    /**
     * Los DNS de la red de salida, como los espera el motor: en la primera línea los
     * servidores y en la segunda los dominios de búsqueda, separados por espacios.
     */
    synchronized String dns() {
        return dns.trim();
    }

    /** Ata un socket del motor a esa red, para que no intente salir por el túnel. */
    boolean atar(int fd) {
        Network red = actual;
        if (red == null) {
            return false;
        }
        try (ParcelFileDescriptor copia = ParcelFileDescriptor.fromFd(fd)) {
            red.bindSocket(copia.getFileDescriptor());
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    @Override
    public synchronized void onAvailable(Network red) {
        activas.put(red, new Datos());
        elegir();
    }

    @Override
    public synchronized void onCapabilitiesChanged(Network red, NetworkCapabilities capacidades) {
        Datos datos = activas.get(red);
        if (datos != null) {
            datos.capacidades = capacidades;
        }
        if (elegir()) {
            publicar();
        }
    }

    @Override
    public synchronized void onLinkPropertiesChanged(Network red, LinkProperties propiedades) {
        Datos datos = activas.get(red);
        if (datos != null) {
            datos.propiedades = propiedades;
        }
        elegir();
        publicar();
    }

    @Override
    public synchronized void onLost(Network red) {
        activas.remove(red);
        elegir();
        publicar();
    }

    /** Vuelve a elegir la red de salida. Devuelve si cambió. */
    private boolean elegir() {
        Network nueva = preferida();
        if (nueva == null ? actual == null : nueva.equals(actual)) {
            return false;
        }
        actual = nueva;
        oyente.cambioLaRed(nueva);
        return true;
    }

    /**
     * Mejor una red que Android ya validó y con DNS; si no hay, se van aflojando las dos
     * condiciones (una red recién conectada sirve antes de que terminen de validarla).
     * Entre iguales, una sin límite de datos (wifi) antes que una medida.
     */
    private Network preferida() {
        for (int paso = 0; paso < 4; paso++) {
            boolean validada = paso < 2, conDns = paso % 2 == 0;
            Network primera = null;
            for (Map.Entry<Network, Datos> entrada : activas.entrySet()) {
                Datos datos = entrada.getValue();
                NetworkCapabilities c = datos.capacidades;
                if (c == null
                    || !c.hasCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
                    || !c.hasCapability(NetworkCapabilities.NET_CAPABILITY_NOT_VPN)
                    || (validada && !c.hasCapability(NetworkCapabilities.NET_CAPABILITY_VALIDATED))
                    || (conDns && (datos.propiedades == null || datos.propiedades.getDnsServers().isEmpty()))) {
                    continue;
                }
                if (c.hasCapability(NetworkCapabilities.NET_CAPABILITY_NOT_METERED)) {
                    return entrada.getKey();
                }
                if (primera == null) {
                    primera = entrada.getKey();
                }
            }
            if (primera != null) {
                return primera;
            }
        }
        return null;
    }

    /** Le cuenta al motor la red de salida (interfaz, puerta de enlace y DNS), si cambió algo. */
    private void publicar() {
        Network red = actual;
        Datos datos = red != null ? activas.get(red) : null;
        if (datos == null || datos.propiedades == null) {
            return; // sin red, o todavía sin sus datos: sigue valiendo lo anterior
        }
        LinkProperties p = datos.propiedades;
        String interfaz = p.getInterfaceName() != null ? p.getInterfaceName() : "";
        String puerta = puertaDeEnlace(p);
        StringBuilder sb = new StringBuilder();
        for (InetAddress servidor : p.getDnsServers()) {
            sb.append(servidor.getHostAddress()).append(' ');
        }
        if (p.getDomains() != null) {
            sb.append('\n').append(p.getDomains());
        }
        String nuevo = interfaz + "|" + puerta + "|" + sb;
        if (nuevo.equals(publicado)) {
            return;
        }
        publicado = nuevo;
        dns = sb.toString();
        Libtailscale.onGatewayChanged(puerta);
        Libtailscale.onDNSConfigChanged(interfaz); // el motor vuelve a leer dns()
    }

    private static String puertaDeEnlace(LinkProperties p) {
        String puerta = "";
        for (RouteInfo ruta : p.getRoutes()) {
            InetAddress ip = ruta.getGateway();
            if (!ruta.isDefaultRoute() || ip == null) {
                continue;
            }
            if (ip instanceof Inet4Address) {
                return ip.getHostAddress(); // la de IPv4 primero
            }
            if (puerta.isEmpty()) {
                puerta = ip.getHostAddress();
            }
        }
        return puerta;
    }
}
