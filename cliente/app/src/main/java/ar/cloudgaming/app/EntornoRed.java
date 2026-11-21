package ar.cloudgaming.app;

import android.content.Context;
import android.os.Build;
import android.provider.Settings;
import android.util.Log;

import org.json.JSONArray;
import org.json.JSONObject;

import java.net.InetAddress;
import java.net.InterfaceAddress;
import java.net.NetworkInterface;
import java.util.Collections;

/**
 * Lo que el motor de red (libtailscale, escrito en Go) le pregunta a Android: dónde
 * guardar su estado, cómo se llama el celu, qué interfaces de red hay.
 *
 * De la app de Tailscale se deja afuera lo de empresas (políticas, certificados propios,
 * claves atadas al hardware) y no se mandan registros a los servidores de Tailscale.
 */
final class EntornoRed implements libtailscale.AppContext {
    private static final String ESTADO = "statestore-";

    private final Context contexto;
    private final Boveda boveda;
    private final VigiaRed vigia;

    EntornoRed(Context contexto, VigiaRed vigia) {
        this.contexto = contexto;
        this.boveda = new Boveda(contexto);
        this.vigia = vigia;
    }

    // ---------------------------------------------------------- estado guardado

    @Override
    public void encryptToPref(String clave, String valor) throws Exception {
        boveda.poner(clave, valor);
    }

    @Override
    public String decryptFromPref(String clave) throws Exception {
        String valor = boveda.sacar(clave);
        return valor != null ? valor : "";
    }

    @Override
    public String getStateStoreKeysJSON() {
        return new JSONArray(boveda.claves(ESTADO)).toString();
    }

    // ------------------------------------------------------------------ el celu

    /** Con este nombre aparece el celu en la red: «cloudgaming-» y el nombre del celu. */
    @Override
    public String getDeviceName() {
        String nombre = Settings.Global.getString(contexto.getContentResolver(), Settings.Global.DEVICE_NAME);
        return "CloudGaming " + (nombre != null && !nombre.isEmpty() ? nombre : Build.MODEL);
    }

    @Override
    public String getOSVersion() {
        return Build.VERSION.RELEASE;
    }

    @Override
    public long getSDKInt() {
        return Build.VERSION.SDK_INT;
    }

    @Override
    public String getInstallSource() {
        return "cloudgaming";
    }

    @Override
    public boolean isChromeOS() {
        return false;
    }

    // ------------------------------------------------------------------- la red

    @Override
    public String getInterfacesAsJson() throws Exception {
        JSONArray lista = new JSONArray();
        java.util.Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
        if (interfaces == null) {
            return lista.toString();
        }
        for (NetworkInterface interfaz : Collections.list(interfaces)) {
            try {
                JSONArray direcciones = new JSONArray();
                for (InterfaceAddress d : interfaz.getInterfaceAddresses()) {
                    InetAddress ip = d.getAddress();
                    if (ip != null && ip.getHostAddress() != null) {
                        direcciones.put(new JSONObject()
                            .put("ip", ip.getHostAddress())
                            .put("prefixLen", d.getNetworkPrefixLength()));
                    }
                }
                lista.put(new JSONObject()
                    .put("name", interfaz.getName())
                    .put("index", interfaz.getIndex())
                    .put("mtu", interfaz.getMTU())
                    .put("up", interfaz.isUp())
                    .put("broadcast", interfaz.supportsMulticast())
                    .put("loopback", interfaz.isLoopback())
                    .put("pointToPoint", interfaz.isPointToPoint())
                    .put("multicast", interfaz.supportsMulticast())
                    .put("addrs", direcciones));
            } catch (Exception e) {
                // Una interfaz que desapareció mientras se la leía: se saltea.
            }
        }
        return lista.toString();
    }

    @Override
    public String getPlatformDNSConfig() {
        return vigia.dns();
    }

    @Override
    public boolean bindSocketToNetwork(int fd) {
        return vigia.atar(fd);
    }

    /** Si Android no informa ningún DNS, usar los de Google antes que quedarse sin nombres. */
    @Override
    public boolean shouldUseGoogleDNSFallback() {
        return true;
    }

    // ----------------------------------------------------------------- registro

    @Override
    public void log(String etiqueta, String linea) {
        Log.d(etiqueta, linea);
    }

    @Override
    public boolean isClientLoggingEnabled() {
        return false;
    }

    // -------------------------------------------------- lo que esta app no usa

    // Políticas de empresa: no hay ninguna. El motor reconoce este mensaje exacto
    // (syspolicy.ErrNoSuchKey) como «esa política no está puesta».
    private static Exception sinPolitica() {
        return new Exception("no such key");
    }

    @Override
    public String getSyspolicyStringValue(String clave) throws Exception {
        throw sinPolitica();
    }

    @Override
    public boolean getSyspolicyBooleanValue(String clave) throws Exception {
        throw sinPolitica();
    }

    @Override
    public String getSyspolicyStringArrayJSONValue(String clave) throws Exception {
        throw sinPolitica();
    }

    @Override
    public byte[] getUserCACertsPEM() {
        return new byte[0];
    }

    @Override
    public boolean hardwareAttestationKeySupported() {
        return false;
    }

    @Override
    public String hardwareAttestationKeyCreate() {
        throw new UnsupportedOperationException();
    }

    @Override
    public void hardwareAttestationKeyLoad(String id) {
        throw new UnsupportedOperationException();
    }

    @Override
    public byte[] hardwareAttestationKeyPublic(String id) {
        throw new UnsupportedOperationException();
    }

    @Override
    public void hardwareAttestationKeyRelease(String id) {
        throw new UnsupportedOperationException();
    }

    @Override
    public byte[] hardwareAttestationKeySign(String id, byte[] datos) {
        throw new UnsupportedOperationException();
    }
}
