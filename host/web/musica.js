'use strict';
/* Música de fondo del menú: ambiental, propia y generada en vivo con Web Audio.
   Un colchón de acordes lentos que respira, notas sueltas tipo campanita que caminan
   por una escala pentatónica, eco y mucha reverb. Como las notas se eligen al azar
   nunca suena igual, y no hay archivos de audio que bajar.

   motorMusica(contexto, salida) sirve tanto para el menú (AudioContext) como para
   armar una muestra en un archivo (OfflineAudioContext). */

const MUSICA = {
  // Notas MIDI (60 = Do central). Re maj9 → Si m11 → Sol maj7(#11) → La sus2.
  acordes: [
    [50, 57, 61, 64, 66],
    [47, 54, 57, 62, 64],
    [43, 50, 54, 57, 61],
    [45, 52, 57, 59, 64],
  ],
  // Re pentatónica (Re Mi Fa# La Si) en el registro de arriba, para las campanitas.
  escala: [74, 76, 78, 81, 83, 86, 88, 90, 93],
  duracion: 9.6, // segundos de cada acorde
  pulso: 0.6,    // grilla de las campanitas
};

function motorMusica(c, salida) {
  const hz = n => 440 * 2 ** ((n - 69) / 12);
  const azar = (min, max) => min + Math.random() * (max - min);

  // Compresor suave al final, para que nada sature.
  const comp = c.createDynamicsCompressor();
  comp.threshold.value = -20;
  comp.ratio.value = 3;
  comp.attack.value = 0.05;
  comp.release.value = 0.4;
  comp.connect(salida);

  // Reverb: la "sala" es ruido que se apaga en 4 s, cada vez más opaco.
  const reverb = c.createConvolver();
  reverb.buffer = respuestaSala(c, 4);
  const vueltaReverb = c.createGain();
  vueltaReverb.gain.value = 0.7;
  reverb.connect(vueltaReverb).connect(comp);

  // Eco de las campanitas (un pulso y medio), que pierde brillo en cada repetición.
  const eco = c.createDelay(2);
  eco.delayTime.value = MUSICA.pulso * 0.75;
  const realimentacion = c.createGain();
  realimentacion.gain.value = 0.38;
  const opaco = c.createBiquadFilter();
  opaco.type = 'lowpass';
  opaco.frequency.value = 2800;
  eco.connect(opaco).connect(realimentacion).connect(eco);
  const vueltaEco = c.createGain();
  vueltaEco.gain.value = 0.45;
  opaco.connect(vueltaEco).connect(comp);
  vueltaEco.connect(reverb);

  // Colchón: sierras desafinadas por un filtro que abre y cierra muy despacio.
  const colchon = c.createGain();
  const filtro = c.createBiquadFilter();
  filtro.type = 'lowpass';
  filtro.frequency.value = 1000;
  filtro.Q.value = 0.6;
  const respiracion = c.createOscillator();
  respiracion.frequency.value = 0.04;
  const profundidad = c.createGain();
  profundidad.gain.value = 450;
  respiracion.connect(profundidad).connect(filtro.frequency);
  respiracion.start();
  colchon.connect(filtro);
  filtro.connect(comp);
  const envioColchon = c.createGain();
  envioColchon.gain.value = 0.6;
  filtro.connect(envioColchon).connect(reverb);

  const campanas = c.createGain();
  campanas.connect(comp);
  campanas.connect(reverb);
  campanas.connect(eco);

  let acorde = 0;
  let paso = 3; // posición de la melodía en la escala

  function acordeLargo(t, notas, dur) {
    for (const n of notas) {
      for (const desafino of [-7, 7]) {
        const o = c.createOscillator();
        o.type = 'sawtooth';
        o.frequency.value = hz(n);
        o.detune.value = desafino;
        const g = c.createGain();
        g.gain.setValueAtTime(0, t);
        g.gain.linearRampToValueAtTime(0.018, t + 2.5);
        g.gain.setValueAtTime(0.018, t + dur);
        g.gain.linearRampToValueAtTime(0, t + dur + 3.5);
        o.connect(g).connect(colchon);
        o.start(t);
        o.stop(t + dur + 3.6);
      }
    }
    // Bajo: la fundamental una octava abajo, sin forma de onda que moleste.
    const b = c.createOscillator();
    b.frequency.value = hz(notas[0] - 12);
    const g = c.createGain();
    g.gain.setValueAtTime(0, t);
    g.gain.linearRampToValueAtTime(0.09, t + 3);
    g.gain.setValueAtTime(0.09, t + dur);
    g.gain.linearRampToValueAtTime(0, t + dur + 3);
    b.connect(g).connect(comp);
    b.start(t);
    b.stop(t + dur + 3.1);
  }

  // Campanita: senoidal con armónicos apenas corridos que se apagan antes que la nota.
  function campana(t, n, vol, pan) {
    const p = c.createStereoPanner();
    p.pan.value = pan;
    p.connect(campanas);
    for (const [mult, peso, caida] of [[1, 1, 3.2], [2.001, 0.35, 1.6], [3.003, 0.12, 0.8], [4.2, 0.05, 0.4]]) {
      const o = c.createOscillator();
      o.frequency.value = hz(n) * mult;
      const g = c.createGain();
      g.gain.setValueAtTime(0.0001, t);
      g.gain.exponentialRampToValueAtTime(vol * peso, t + 0.008);
      g.gain.exponentialRampToValueAtTime(0.0001, t + caida);
      o.connect(g).connect(p);
      o.start(t);
      o.stop(t + caida + 0.05);
    }
  }

  // Un acorde entero con su frase de campanitas, a partir del segundo t del contexto.
  function bloque(t) {
    acordeLargo(t, MUSICA.acordes[acorde++ % MUSICA.acordes.length], MUSICA.duracion);
    const pasos = Math.round(MUSICA.duracion / MUSICA.pulso);
    for (let i = 0; i < pasos; i++) {
      if (Math.random() > 0.3) continue; // silencios: la frase respira
      paso = Math.min(Math.max(paso + [-2, -1, -1, 1, 1, 2][Math.floor(Math.random() * 6)], 0), MUSICA.escala.length - 1);
      const tn = t + i * MUSICA.pulso + azar(0, 0.03); // un poco "a mano"
      campana(tn, MUSICA.escala[paso], azar(0.05, 0.09), azar(-0.6, 0.6));
      if (Math.random() < 0.15 && paso >= 2) {
        campana(tn + 0.02, MUSICA.escala[paso - 2], azar(0.03, 0.05), azar(-0.6, 0.6));
      }
    }
  }

  return { bloque };
}

function respuestaSala(c, segundos) {
  const largo = Math.round(c.sampleRate * segundos);
  const buffer = c.createBuffer(2, largo, c.sampleRate);
  for (let canal = 0; canal < 2; canal++) {
    const datos = buffer.getChannelData(canal);
    let filtrado = 0;
    for (let i = 0; i < largo; i++) {
      const avance = i / largo;
      // Pasa-bajos de un polo que se cierra con el tiempo: la cola queda más opaca.
      filtrado += (Math.random() * 2 - 1 - filtrado) * (0.6 - 0.5 * avance);
      datos[i] = filtrado * (1 - avance) ** 2.5;
    }
  }
  return buffer;
}
