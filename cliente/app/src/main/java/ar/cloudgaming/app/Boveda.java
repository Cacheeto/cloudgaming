package ar.cloudgaming.app;

import android.content.Context;
import android.content.SharedPreferences;
import android.security.keystore.KeyGenParameterSpec;
import android.security.keystore.KeyProperties;
import android.util.Base64;

import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.InvalidKeyException;
import java.security.KeyStore;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import javax.crypto.AEADBadTagException;
import javax.crypto.Cipher;
import javax.crypto.KeyGenerator;
import javax.crypto.SecretKey;
import javax.crypto.spec.GCMParameterSpec;
import javax.crypto.spec.SecretKeySpec;

/**
 * Guarda cifrado lo que la red necesita recordar; sobre todo, la identidad de este celu
 * (su clave privada). Los valores van con AES-GCM. La llave que los cifra se guarda
 * envuelta por otra que nunca sale del almacén de claves de Android.
 */
final class Boveda {
    private static final String ALIAS = "cloudgaming-red";
    private static final String LLAVE = "_llave";
    private static final String AES_GCM = "AES/GCM/NoPadding";
    private static final int LARGO_IV = 12;

    private final SharedPreferences prefs;
    private SecretKey llave;

    Boveda(Context contexto) {
        prefs = contexto.getSharedPreferences("red", Context.MODE_PRIVATE);
    }

    synchronized void poner(String clave, String valor) throws GeneralSecurityException, IOException {
        String sellado = sellar(llave(), valor.getBytes(StandardCharsets.UTF_8));
        if (!prefs.edit().putString(clave, sellado).commit()) {
            throw new IOException("no se pudo guardar " + clave);
        }
    }

    /** El valor guardado, o null si no hay. */
    synchronized String sacar(String clave) throws GeneralSecurityException, IOException {
        SecretKey llave = llave(); // antes de leer: si la llave se perdió, vacía lo guardado
        String sellado = prefs.getString(clave, null);
        return sellado == null ? null : new String(abrir(llave, sellado), StandardCharsets.UTF_8);
    }

    /** Las claves guardadas que empiezan con el prefijo, sin el prefijo. */
    synchronized List<String> claves(String prefijo) {
        List<String> claves = new ArrayList<>();
        for (String clave : prefs.getAll().keySet()) {
            if (clave.startsWith(prefijo)) {
                claves.add(clave.substring(prefijo.length()));
            }
        }
        return claves;
    }

    private SecretKey llave() throws GeneralSecurityException, IOException {
        if (llave != null) {
            return llave;
        }
        KeyStore almacen = KeyStore.getInstance("AndroidKeyStore");
        almacen.load(null);
        SecretKey maestra = (SecretKey) almacen.getKey(ALIAS, null);
        String envuelta = prefs.getString(LLAVE, null);
        if (maestra != null && envuelta != null) {
            try {
                llave = new SecretKeySpec(abrir(maestra, envuelta), "AES");
                return llave;
            } catch (AEADBadTagException | InvalidKeyException e) {
                // La llave maestra ya no es la que envolvió a esta (o Android la invalidó):
                // lo guardado no se puede leer más. Se empieza de cero y hay que iniciar sesión.
                almacen.deleteEntry(ALIAS);
                maestra = null;
            }
        }
        if (maestra == null) {
            KeyGenerator generador = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore");
            generador.init(new KeyGenParameterSpec.Builder(ALIAS,
                KeyProperties.PURPOSE_ENCRYPT | KeyProperties.PURPOSE_DECRYPT)
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setKeySize(256)
                .build());
            maestra = generador.generateKey();
        }
        byte[] nueva = new byte[32];
        new SecureRandom().nextBytes(nueva);
        if (!prefs.edit().clear().putString(LLAVE, sellar(maestra, nueva)).commit()) {
            throw new IOException("no se pudo guardar la llave");
        }
        llave = new SecretKeySpec(nueva, "AES");
        return llave;
    }

    /** IV (al azar, lo elige el cifrador) + texto cifrado, en base64. */
    private static String sellar(SecretKey llave, byte[] claro) throws GeneralSecurityException {
        Cipher cifrador = Cipher.getInstance(AES_GCM);
        cifrador.init(Cipher.ENCRYPT_MODE, llave);
        byte[] iv = cifrador.getIV();
        byte[] cifrado = cifrador.doFinal(claro);
        if (iv == null || iv.length != LARGO_IV) {
            throw new GeneralSecurityException("IV inesperado");
        }
        byte[] todo = Arrays.copyOf(iv, iv.length + cifrado.length);
        System.arraycopy(cifrado, 0, todo, iv.length, cifrado.length);
        return Base64.encodeToString(todo, Base64.NO_WRAP);
    }

    private static byte[] abrir(SecretKey llave, String sellado) throws GeneralSecurityException {
        byte[] todo;
        try {
            todo = Base64.decode(sellado, Base64.NO_WRAP);
        } catch (IllegalArgumentException e) {
            throw new AEADBadTagException("no es base64");
        }
        if (todo.length < LARGO_IV) {
            throw new AEADBadTagException("valor cortado");
        }
        Cipher cifrador = Cipher.getInstance(AES_GCM);
        cifrador.init(Cipher.DECRYPT_MODE, llave, new GCMParameterSpec(128, todo, 0, LARGO_IV));
        return cifrador.doFinal(todo, LARGO_IV, todo.length - LARGO_IV);
    }
}
