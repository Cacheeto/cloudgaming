'use strict';
/* Panel CloudGaming: menú de barra cruzada para manejar Sunshine y WoW desde el celu. */

const $ = (sel, raiz = document) => raiz.querySelector(sel);
const limitar = (n, min, max) => Math.min(Math.max(n, min), max);

function crear(etiqueta, clase, texto) {
  const el = document.createElement(etiqueta);
  if (clase) el.className = clase;
  if (texto != null) el.textContent = texto;
  return el;
}

// ------------------------------------------------------------------ íconos

const ICONOS = {
  pulso: '<path d="M2.5 12h4l2.5-6.5 4 13 3-9.5 1.5 3h4"/>',
  mando: '<path d="M6.5 7.5h11a4.5 4.5 0 0 1 4.4 5.5l-.8 3.6a2.4 2.4 0 0 1-4.2 1l-1.7-2.1H8.8l-1.7 2.1a2.4 2.4 0 0 1-4.2-1L2.1 13a4.5 4.5 0 0 1 4.4-5.5z"/><path d="M8 10v3.6M6.2 11.8h3.6"/><circle cx="15.6" cy="10.8" r=".9"/><circle cx="17.8" cy="13" r=".9"/>',
  senal: '<path d="M4.5 19.5v-3M9.5 19.5V13M14.5 19.5v-10M19.5 19.5v-14"/>',
  celu: '<rect x="6.5" y="2.5" width="11" height="19" rx="2.5"/><path d="M10.5 18.5h3"/>',
  pc: '<rect x="3" y="4" width="18" height="12" rx="1.5"/><path d="M12 16v4M8.5 20h7"/>',
  chip: '<rect x="6" y="6" width="12" height="12" rx="2"/><rect x="9.5" y="9.5" width="5" height="5" rx=".8"/><path d="M9 2.5V6M15 2.5V6M9 18v3.5M15 18v3.5M2.5 9H6M2.5 15H6M18 9h3.5M18 15h3.5"/>',
  ajustes: '<path d="M4 7h9M17 7h3M4 12h3M11 12h9M4 17h11M19 17h1"/><circle cx="15" cy="7" r="2"/><circle cx="9" cy="12" r="2"/><circle cx="17" cy="17" r="2"/>',
  sol: '<circle cx="12" cy="12" r="4"/><path d="M12 2.5v2M12 19.5v2M4.6 4.6 6 6M18 18l1.4 1.4M2.5 12h2M19.5 12h2M4.6 19.4 6 18M18 6l1.4-1.4"/>',
  tele: '<rect x="2.5" y="4.5" width="19" height="12.5" rx="2"/><path d="M8 21h8M12 17v4"/>',
  espada: '<path d="M20 4h-4.5L6.5 13l4.5 4.5L20 8.5z"/><path d="M5 15.5 8.5 19M3.5 20.5l3-3"/>',
  play: '<path d="M8 5.5v13l10.5-6.5z"/>',
  refrescar: '<path d="M20 11.5a8 8 0 1 1-2.6-5.9"/><path d="M20.5 4v4.5H16"/>',
  apagar: '<path d="M12 3v8"/><path d="M6.6 6.3a8 8 0 1 0 10.8 0"/>',
  cerrar: '<path d="M6 6l12 12M18 6 6 18"/>',
  wifi: '<path d="M2.5 9a14 14 0 0 1 19 0M5.5 12.5a9.5 9.5 0 0 1 13 0M8.7 16a5 5 0 0 1 6.6 0"/><circle cx="12" cy="19.3" r=".9"/>',
  antena: '<path d="M12 10v11"/><path d="M8.5 6.5a5 5 0 0 0 0 7M15.5 6.5a5 5 0 0 1 0 7M5.5 3.5a9.5 9.5 0 0 0 0 13M18.5 3.5a9.5 9.5 0 0 1 0 13"/><circle cx="12" cy="10" r="1.2"/>',
  nube: '<path d="M7 18.5h10a4 4 0 0 0 .6-8 6 6 0 0 0-11.5 1.6A3.3 3.3 0 0 0 7 18.5z"/>',
  calendario: '<rect x="3.5" y="5" width="17" height="15.5" rx="2"/><path d="M3.5 10h17M8 3v4M16 3v4"/>',
  reloj: '<circle cx="12" cy="12" r="8.5"/><path d="M12 7.5V12l3 2"/>',
  escudo: '<path d="M12 3 4.5 6v5.5c0 4.5 3.2 8.2 7.5 9.5 4.3-1.3 7.5-5 7.5-9.5V6z"/><path d="M9 12l2 2 4-4.5"/>',
  barras: '<path d="M3.5 20.5h17"/><rect x="5.5" y="12" width="3" height="6" rx=".6"/><rect x="10.5" y="7" width="3" height="11" rx=".6"/><rect x="15.5" y="10" width="3" height="8" rx=".6"/>',
  borrar: '<path d="M4.5 7h15M9.5 7V4.5h5V7M6.5 7l1 13h9l1-13"/>',
  info: '<circle cx="12" cy="12" r="8.5"/><path d="M12 11v5.5"/><circle cx="12" cy="7.8" r=".9"/>',
  mas: '<circle cx="12" cy="12" r="8.5"/><path d="M12 8v8M8 12h8"/>',
  lista: '<path d="M8.5 6.5h11M8.5 12h11M8.5 17.5h11"/><circle cx="4.8" cy="6.5" r=".9"/><circle cx="4.8" cy="12" r=".9"/><circle cx="4.8" cy="17.5" r=".9"/>',
  memoria: '<rect x="3" y="7" width="18" height="10" rx="1.5"/><path d="M7 10.5v3M11 10.5v3M15 10.5v3M6 17v2.5M10 17v2.5M14 17v2.5M18 17v2.5"/>',
  paleta: '<path d="M12 3.5a8.5 8.5 0 0 0 0 17c1.3 0 1.9-.8 1.9-1.7 0-1.2-1-1.5-1-2.6 0-1 .8-1.7 1.9-1.7h2.3a3.4 3.4 0 0 0 3.4-3.4C20.5 6.9 16.7 3.5 12 3.5z"/><circle cx="7.8" cy="11" r="1"/><circle cx="10.5" cy="7.5" r="1"/><circle cx="15" cy="7.8" r="1"/>',
  ondas: '<path d="M2.5 9c3-3 6 3 9.5 0s6.5 3 9.5 0M2.5 15c3-3 6 3 9.5 0s6.5 3 9.5 0"/>',
  sonido: '<path d="M4 9.5h3.5L12 5.5v13l-4.5-4H4z"/><path d="M15.5 9a4 4 0 0 1 0 6M18 6.5a7.5 7.5 0 0 1 0 11"/>',
  expandir: '<path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"/>',
  salir: '<path d="M14 4.5h4.5A1.5 1.5 0 0 1 20 6v12a1.5 1.5 0 0 1-1.5 1.5H14"/><path d="M10 8l-4 4 4 4M6 12h9"/>',
  usuario: '<circle cx="12" cy="8.5" r="4"/><path d="M4.5 20.5a7.5 7.5 0 0 1 15 0"/>',
  desconectar: '<path d="M9.5 14.5 7 17a3.5 3.5 0 0 1-5-5l2.5-2.5M14.5 9.5 17 7a3.5 3.5 0 0 1 5 5l-2.5 2.5M8 8 5 5M16 16l3 3"/>',
  campana: '<path d="M6 16.5V11a6 6 0 0 1 12 0v5.5l1.5 2h-15z"/><path d="M10 20.5a2 2 0 0 0 4 0"/>',
  musica: '<path d="M9 18.5V6.5l10.5-2.5v12"/><circle cx="6.5" cy="18.5" r="2.5"/><circle cx="17" cy="16" r="2.5"/>',
  teclado: '<rect x="2.5" y="6" width="19" height="12.5" rx="2"/><path d="M8 15h8"/><circle cx="6.5" cy="9.8" r=".9"/><circle cx="10.2" cy="9.8" r=".9"/><circle cx="13.8" cy="9.8" r=".9"/><circle cx="17.5" cy="9.8" r=".9"/><circle cx="6.5" cy="12.6" r=".9"/><circle cx="17.5" cy="12.6" r=".9"/>',
};

const icono = nombre => `<svg viewBox="0 0 24 24" aria-hidden="true">${ICONOS[nombre] || ''}</svg>`;

// ------------------------------------------------------------- preferencias

const PREFS_DEFECTO = { tema: 'noche', ondas: true, sonido: false, musica: 'media', ayudaVista: false };

function cargarPrefs() {
  try {
    return { ...PREFS_DEFECTO, ...JSON.parse(localStorage.getItem('panel-prefs') || '{}') };
  } catch {
    return { ...PREFS_DEFECTO };
  }
}

function guardarPrefs() {
  try { localStorage.setItem('panel-prefs', JSON.stringify(S.prefs)); } catch { /* sin almacenamiento */ }
}

const TEMAS = [
  { id: 'noche', nombre: 'Azul noche', h: 222 },
  { id: 'violeta', nombre: 'Violeta', h: 272 },
  { id: 'esmeralda', nombre: 'Esmeralda', h: 162 },
  { id: 'rubi', nombre: 'Rubí', h: 350 },
  { id: 'ambar', nombre: 'Ámbar', h: 32 },
  { id: 'hora', nombre: 'Según la hora del día', h: null },
];

function tonoSegunHora(d = new Date()) {
  const puntos = [[0, 236], [5, 262], [7, 300], [9, 205], [13, 192], [17, 28], [19.5, 300], [21.5, 248], [24, 236]];
  const hora = d.getHours() + d.getMinutes() / 60;
  for (let i = 1; i < puntos.length; i++) {
    const [h0, t0] = puntos[i - 1], [h1, t1] = puntos[i];
    if (hora <= h1) return Math.round(t0 + (t1 - t0) * (hora - h0) / (h1 - h0));
  }
  return 236;
}

function hslAHex(h, s, l) {
  s /= 100; l /= 100;
  const f = n => {
    const k = (n + h / 30) % 12;
    const c = l - s * Math.min(l, 1 - l) * Math.max(-1, Math.min(k - 3, 9 - k, 1));
    return Math.round(c * 255).toString(16).padStart(2, '0');
  };
  return `#${f(0)}${f(8)}${f(4)}`;
}

function aplicarTema() {
  const tema = TEMAS.find(t => t.id === S.prefs.tema) || TEMAS[0];
  const h = tema.h ?? tonoSegunHora();
  document.documentElement.style.setProperty('--h', h);
  $('meta[name=theme-color]').content = hslAHex(h, 70, 9);
}

// ------------------------------------------------------------------- estado

const S = {
  cat: 0,
  sel: {},
  estado: null,
  errorEstado: null,
  clientes: null,
  prefs: cargarPrefs(),
  entrada: 'toque',
  pasos: 0,
  // Entró por ser un equipo de la misma cuenta de la red: no hay sesión que cerrar.
  porCuenta: false,
};

// Dentro de la app CloudGaming para Android existe el objeto CloudGaming.
// Desde la 2.0 la app trae el stream adentro (código de Moonlight incluido) y desde la 2.1
// sus ajustes se manejan desde este menú (ajustesStream / guardarAjuste).
const EN_APP = typeof window.CloudGaming !== 'undefined';
const STREAM_PROPIO = EN_APP && typeof window.CloudGaming.ajustesStream === 'function';
// Desde la 3.0 la app trae la red adentro (el motor de Tailscale, sin su app) y cuenta cómo
// está: 'lista', 'conectando' o lo que le falta. Se conecta sola, no hay nada que prender.
const RED_PROPIA = EN_APP && typeof window.CloudGaming.red === 'function';
// De la 2.3 a la 2.5 dependía de la app de Tailscale y solo sabía si había una VPN prendida.
const sinRed = () => (RED_PROPIA ? window.CloudGaming.red() !== 'lista'
  : EN_APP && typeof window.CloudGaming.vpnActiva === 'function' && !window.CloudGaming.vpnActiva());
const EN_ANDROID = /Android/i.test(navigator.userAgent);
const VERSION_APP = '3.6';
// El APK de la versión instalada: el del celu (64 bits) o el de la tele (32 bits, app 3.2+).
const APK = EN_APP && typeof window.CloudGaming.apk === 'function' ? window.CloudGaming.apk() : 'cloudgaming.apk';
// Desde la 3.3 la app baja la actualización por su propia red. Hasta la 3.2 la bajaba el
// navegador, que desde la 3.0 no llega a la PC (la red es solo de la app).
const SE_ACTUALIZA_SOLA = EN_APP && typeof window.CloudGaming.actualizar === 'function';
// Desde la 3.4 la app busca sola las PC de la red (no trae una dirección fija) y se puede cambiar.
const CAMBIA_DE_PC = EN_APP && typeof window.CloudGaming.cambiarPc === 'function';
const PAQUETE_APP = 'ar.cloudgaming.app';

// ----------------------------------------------------------------- formatos

const NF0 = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 0 });
const NF1 = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 1 });
const NF2 = new Intl.NumberFormat('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

function tamano(n) {
  if (n == null) return '—';
  if (n < 1e6) return `${NF0.format(n / 1e3)} KB`;
  if (n < 1e9) return `${NF0.format(n / 1e6)} MB`;
  return `${NF2.format(n / 1e9)} GB`;
}

const tamanoBin = n => (n == null ? '—' : `${NF1.format(n / 2 ** 30)} GB`);
const gb = n => `${NF1.format(n)} GB`;

function horaCorta(texto) {
  if (!texto) return '—';
  const d = new Date(texto.replace(' ', 'T'));
  return d.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' });
}

function fechaCorta(dia) {
  const [, m, d] = dia.split('-');
  return `${+d}/${+m}`;
}

function duracion(seg) {
  const d = Math.floor(seg / 86400), h = Math.floor(seg % 86400 / 3600), m = Math.floor(seg % 3600 / 60);
  if (d) return `${d} d ${h} h`;
  if (h) return `${h} h ${m} min`;
  return `${m} min`;
}

function hace(texto) {
  const seg = (Date.now() - new Date(texto)) / 1000;
  if (seg < 120) return 'recién';
  if (seg < 3600) return `hace ${Math.round(seg / 60)} min`;
  if (seg < 86400) return `hace ${Math.round(seg / 3600)} h`;
  const dias = Math.round(seg / 86400);
  return dias === 1 ? 'ayer' : `hace ${dias} días`;
}

const MENSAJES = {
  sin_conexion: 'No llego a la PC. ¿Está prendida y con Tailscale?',
  sunshine_apagado: 'Sunshine está apagado',
  credenciales_sunshine: 'Sunshine rechazó el usuario o la contraseña',
  demasiados_intentos: 'Demasiados intentos: esperá un minuto',
  sesion: 'Tu sesión venció: entrá de nuevo',
  datos_invalidos: 'Algún dato no es válido',
};

function mensajeError(e) {
  const codigo = e?.codigo || e;
  if (codigo === 'sin_conexion' && sinRed()) {
    return RED_PROPIA ? 'Todavía conectando con la PC…' : 'Tailscale está apagado en el celu';
  }
  return MENSAJES[codigo] || `Algo falló (${codigo})`;
}

// --------------------------------------------------------------------- API

class ErrorApi extends Error {
  constructor(codigo) {
    super(codigo);
    this.codigo = codigo;
  }
}

// Tope de cada pedido: si la red de la app se cae a mitad de uno, el pedido no queda colgado
// minutos (y con él los que vienen atrás, que esperan conexión libre).
const ESPERA_API_MS = 20000;

async function api(ruta, cuerpo) {
  const corte = new AbortController();
  const reloj = setTimeout(() => corte.abort(), ESPERA_API_MS);
  try {
    return await pedirApi(ruta, cuerpo, corte.signal);
  } finally {
    clearTimeout(reloj);
  }
}

async function pedirApi(ruta, cuerpo, signal) {
  const opciones = { headers: {}, credentials: 'same-origin', signal };
  if (cuerpo !== undefined) {
    opciones.method = 'POST';
    opciones.headers['Content-Type'] = 'application/json';
    opciones.headers['X-Panel'] = '1';
    opciones.body = JSON.stringify(cuerpo);
  }
  let r;
  try {
    r = await fetch(`api/${ruta}`, opciones);
  } catch {
    throw new ErrorApi('sin_conexion');
  }
  let dato = null;
  try { dato = await r.json(); } catch { /* respuesta vacía */ }
  if (r.status === 401) {
    mostrarEntrada(MENSAJES.sesion);
    throw new ErrorApi('sesion');
  }
  if (!r.ok) throw new ErrorApi(dato?.error || `http_${r.status}`);
  return dato;
}

// ------------------------------------------------------------------ sonidos

let audio = null;
const SONIDOS = {
  item: [1320, 0.025, 0.03], cat: [880, 0.045, 0.045], ok: [660, 0.07, 0.05],
  abrir: [990, 0.05, 0.035], atras: [520, 0.05, 0.035], nada: [210, 0.07, 0.05],
};

function sonar(tipo) {
  if (!S.prefs.sonido) return;
  if (navigator.vibrate && (tipo === 'item' || tipo === 'cat')) navigator.vibrate(tipo === 'cat' ? 8 : 4);
  const s = SONIDOS[tipo];
  if (!s) return;
  try {
    audio ??= new (window.AudioContext || window.webkitAudioContext)();
    const [frec, dur, vol] = s;
    const t = audio.currentTime;
    const osc = audio.createOscillator();
    const gan = audio.createGain();
    osc.type = 'sine';
    osc.frequency.setValueAtTime(frec, t);
    if (tipo === 'ok') osc.frequency.exponentialRampToValueAtTime(frec * 1.5, t + dur);
    gan.gain.setValueAtTime(0.0001, t);
    gan.gain.exponentialRampToValueAtTime(vol, t + 0.005);
    gan.gain.exponentialRampToValueAtTime(0.0001, t + dur);
    osc.connect(gan).connect(audio.destination);
    osc.start(t);
    osc.stop(t + dur + 0.02);
  } catch { /* sin audio */ }
}

// ------------------------------------------------------------------- música
// El motor está en musica.js. Acá se decide cuándo suena: con la preferencia prendida,
// la página a la vista y sin un juego arrancando. El navegador no deja empezar a sonar
// hasta el primer toque o botón, así que también se intenta con cada uno.

const NIVELES_MUSICA = { no: 0, baja: 0.25, media: 0.5, alta: 1 };
const NOMBRES_MUSICA = { no: 'Apagada', baja: 'Baja', media: 'Media', alta: 'Alta' };

const Musica = {
  c: null,
  volumen: null,
  motor: null,
  siguiente: 0,
  timer: 0,
  sonando: false,
  enJuego: false,
  enPausa: false,

  iniciar() {
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'visible') this.enJuego = false; // volviste del juego
      this.actualizar();
    });
    // En captura: corre antes que el menú, que puede estar por lanzar un juego.
    for (const tipo of ['pointerdown', 'keydown']) {
      addEventListener(tipo, () => {
        this.enJuego = false;
        this.actualizar();
      }, { capture: true, passive: true });
    }
    this.actualizar();
  },

  nivel() {
    return NIVELES_MUSICA[S.prefs.musica] ?? 0;
  },

  actualizar() {
    if (this.nivel() > 0 && document.visibilityState === 'visible' && !this.enJuego && !this.enPausa) this.sonar();
    else this.callar();
  },

  // Al tocar un juego, la música se va antes de que arranque el stream.
  alJugar() {
    this.enJuego = true;
    this.actualizar();
  },

  sonar() {
    try {
      if (!this.c) {
        this.c = new (window.AudioContext || window.webkitAudioContext)();
        this.volumen = this.c.createGain();
        this.volumen.gain.value = 0;
        this.volumen.connect(this.c.destination);
        this.motor = motorMusica(this.c, this.volumen);
        this.siguiente = this.c.currentTime + 0.3;
      }
      if (this.c.state === 'running') this.subir();
      else this.c.resume().then(() => this.subir(), () => { /* todavía sin permiso */ });
    } catch { /* sin audio */ }
  },

  subir() {
    if (this.nivel() === 0 || this.enJuego || this.enPausa) return; // cambió mientras esperaba el permiso
    const t = this.c.currentTime;
    this.volumen.gain.cancelScheduledValues(t);
    this.volumen.gain.setTargetAtTime(this.nivel(), t, 1.5); // entra despacio
    if (!this.sonando) {
      this.sonando = true;
      this.programar();
    }
  },

  // Arma cada acorde un poco antes de que suene.
  programar() {
    const t = this.c.currentTime;
    if (this.siguiente < t) this.siguiente = t + 0.1;
    while (this.siguiente < t + 1.5) {
      this.motor.bloque(this.siguiente);
      this.siguiente += MUSICA.duracion;
    }
    this.timer = setTimeout(() => this.programar(), 500);
  },

  callar() {
    if (!this.c || !this.sonando) return;
    this.sonando = false;
    clearTimeout(this.timer);
    const t = this.c.currentTime;
    this.volumen.gain.cancelScheduledValues(t);
    this.volumen.gain.setTargetAtTime(0, t, 0.25);
    // Pasado el fundido se suspende: no gasta batería mientras no suena.
    setTimeout(() => { if (!this.sonando) this.c.suspend(); }, 1500);
  },
};

// -------------------------------------------------------------------- ondas

const Ondas = {
  raf: 0,
  ultimo: 0,

  iniciar() {
    this.lienzo = $('#ondas');
    this.ctx = this.lienzo.getContext('2d');
    addEventListener('resize', () => this.medir());
    document.addEventListener('visibilitychange', () => this.actualizar());
    this.medir();
    this.actualizar();
  },

  medir() {
    const dpr = Math.min(devicePixelRatio || 1, 1.5);
    this.w = innerWidth;
    this.h = innerHeight;
    this.lienzo.width = Math.round(this.w * dpr);
    this.lienzo.height = Math.round(this.h * dpr);
    this.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    this.dibujar(performance.now());
  },

  enMovimiento() {
    return S.prefs.ondas && document.visibilityState === 'visible'
      && !matchMedia('(prefers-reduced-motion: reduce)').matches;
  },

  actualizar() {
    cancelAnimationFrame(this.raf);
    if (this.enMovimiento()) this.raf = requestAnimationFrame(t => this.cuadro(t));
    else this.dibujar(performance.now());
  },

  cuadro(t) {
    if (t - this.ultimo > 33) { // ~30 cuadros por segundo alcanzan y cuidan la batería
      this.ultimo = t;
      this.dibujar(t);
    }
    if (this.enMovimiento()) this.raf = requestAnimationFrame(t2 => this.cuadro(t2));
  },

  dibujar(t) {
    const { ctx, w, h } = this;
    const s = t / 1000;
    const base = h * (h > w ? 0.76 : 0.64);
    ctx.clearRect(0, 0, w, h);
    ctx.globalCompositeOperation = 'lighter';
    const capas = [
      { a: h * 0.065, l: w * 1.1, v: 0.12, g: 26, op: 0.09, f: 0 },
      { a: h * 0.045, l: w * 0.8, v: -0.09, g: 14, op: 0.08, f: 1.7 },
      { a: h * 0.085, l: w * 1.45, v: 0.07, g: 42, op: 0.055, f: 3.1 },
    ];
    for (const c of capas) {
      const y = (px, desfase) => base
        + Math.sin(px / c.l * Math.PI * 2 + s * c.v * 6 + c.f + desfase) * c.a
        + Math.sin(px / (c.l * 0.45) + s * c.v * 3.3 + c.f) * c.a * 0.35;
      ctx.beginPath();
      for (let px = 0; px <= w + 16; px += 16) {
        const yy = y(px, 0);
        if (px) ctx.lineTo(px, yy); else ctx.moveTo(px, yy);
      }
      for (let px = w + 16; px >= 0; px -= 16) {
        const grosor = c.g * (0.6 + 0.4 * Math.sin(px / w * 3 + s * 0.4 + c.f));
        ctx.lineTo(px, y(px, 0.35) + grosor);
      }
      ctx.closePath();
      const grad = ctx.createLinearGradient(0, base - c.a, 0, base + c.a + c.g);
      grad.addColorStop(0, `rgba(255,255,255,${c.op * 1.6})`);
      grad.addColorStop(1, 'rgba(255,255,255,0)');
      ctx.fillStyle = grad;
      ctx.fill();
      ctx.beginPath();
      for (let px = 0; px <= w + 16; px += 16) {
        const yy = y(px, 0);
        if (px) ctx.lineTo(px, yy); else ctx.moveTo(px, yy);
      }
      ctx.strokeStyle = `rgba(255,255,255,${c.op * 2})`;
      ctx.lineWidth = 1;
      ctx.stroke();
    }
    ctx.globalCompositeOperation = 'source-over';
  },
};

// ------------------------------------------------------------ textos útiles

const VIAS = {
  casa: { texto: 'En el wifi de casa', icono: 'wifi', tono: 'ok', corto: 'Casa' },
  datos: { texto: 'Fuera de casa · conexión directa', icono: 'antena', tono: 'ok', corto: 'Datos' },
  rebote: { texto: 'Fuera de casa · por servidor de rebote (más lento)', icono: 'nube', tono: 'alerta', corto: 'Rebote' },
  inactivo: { texto: 'Conectado a Tailscale, sin tráfico', icono: 'celu', tono: '', corto: 'En línea' },
};

function viaCelu(celu) {
  if (!celu) return { texto: 'No encuentro el celu en Tailscale', icono: 'celu', tono: 'error', corto: 'Sin celu' };
  if (!celu.online) return { texto: 'El celu está desconectado de la red', icono: 'celu', tono: 'alerta', corto: 'Desconectado' };
  return VIAS[celu.via] || VIAS.inactivo;
}

function textoJuego(j) {
  if (j.wow) return 'Abierto';
  if (j.battlenet) return 'Battle.net abierto · juego cerrado';
  return 'Cerrado';
}

const textoCiclo = d => (d.limite ? `${tamano(d.ciclo)} de ${gb(d.config.limite_gb)}` : `${tamano(d.ciclo)} · sin límite`);

function tonoDatos(d) {
  if (d.pct == null) return '';
  if (d.pct >= 100) return 'error';
  if (d.pct >= d.config.aviso_pct) return 'alerta';
  return '';
}

function sinEstado() {
  if (S.errorEstado === 'sin_conexion' && sinRed()) {
    return [RED_PROPIA ? {
      id: 'red', icono: 'antena', titulo: 'Conectando con la PC…', tono: 'alerta',
      sub: 'El menú vuelve solo · si tarda, revisá que el celu tenga internet', accion: actualizar,
    } : {
      id: 'tailscale', icono: 'antena', titulo: 'Tailscale está apagado', tono: 'error',
      sub: 'Tocá para abrirlo y prenderlo · el menú vuelve solo', accion: () => window.CloudGaming.abrirTailscale(),
    }];
  }
  return [{
    id: 'cargando', icono: S.errorEstado ? 'info' : 'reloj',
    titulo: S.errorEstado ? 'Sin conexión con la PC' : 'Conectando…',
    sub: S.errorEstado ? mensajeError(S.errorEstado) : '',
    tono: S.errorEstado ? 'error' : '',
    accion: S.errorEstado ? actualizar : null,
  }];
}

// --------------------------------------------------------------- categorías

const esWow = nombre => /wow|warcraft/i.test(nombre);

function iconoApp(nombre) {
  if (esWow(nombre)) return 'espada';
  if (/steam/i.test(nombre)) return 'mando';
  return 'tele';
}

function itemsJugar() {
  const e = S.estado;
  if (!e) return sinEstado();
  const sun = e.sunshine;
  const apps = e.moonlight?.apps || [];
  // Recientes: las últimas jugadas (la PC las manda ya ordenadas). Actuales: el resto, WoW arriba.
  const recientes = (e.recientes || []).filter(r => apps.includes(r.app)).slice(0, 3);
  const actuales = apps.filter(a => !recientes.some(r => r.app === a)).sort((a, b) => esWow(b) - esWow(a));
  const item = (nombre, cuando, grupo) => {
    const enCurso = sun.ocupado && sun.app === nombre;
    return {
      id: `jugar-${nombre}`, icono: iconoApp(nombre), titulo: nombre, grupo,
      sub: enCurso ? 'En curso · tocá para volver' : sun.ocupado ? 'Hay otra app abierta en la PC'
        : cuando ? `Jugado ${hace(cuando)}` : 'Tocá para jugar',
      tono: enCurso ? 'ok' : '',
      accion: () => jugar(nombre),
    };
  };
  return [
    ...recientes.map((r, i) => item(r.app, r.cuando, i === 0 ? 'Recientes' : null)),
    ...actuales.map((a, i) => item(a, null, i === 0 ? 'Actuales' : null)),
  ];
}

function itemsEstado() {
  const e = S.estado;
  if (!e) return sinEstado();
  const sun = e.sunshine, j = e.juego, d = e.datos, via = viaCelu(d.celu);
  return [
    {
      id: 'sunshine', icono: 'sol', titulo: 'Sunshine',
      sub: !sun.activo ? 'Apagado' : sun.ocupado ? 'Transmitiendo' : 'Listo para jugar',
      tono: sun.activo ? 'ok' : 'error', accion: hojaSunshine,
    },
    {
      id: 'sesion', icono: 'tele', titulo: sun.ocupado ? (sun.app || 'App abierta') : 'Sin sesión',
      sub: sun.ocupado
        ? `Desde las ${horaCorta(sun.desde)}${sun.cliente ? '' : ' · sin nadie conectado'}`
        : 'Elegí un juego en «Jugar»',
      tono: sun.ocupado && !sun.cliente ? 'alerta' : '',
      accion: sun.ocupado ? confirmarTerminar : null,
    },
    { id: 'celu', icono: via.icono, titulo: d.celu?.nombre || 'Celu', sub: via.texto, tono: via.tono, accion: hojaCelu },
    // WoW solo aparece mientras está abierto; cerrado no aporta nada acá.
    j?.wow ? { id: 'juego', icono: 'espada', titulo: 'World of Warcraft', sub: textoJuego(j), tono: 'ok', accion: () => irA('juego') } : null,
    { id: 'datos', icono: 'senal', titulo: 'Datos fuera de casa', sub: textoCiclo(d), tono: tonoDatos(d), pct: d.pct, accion: () => irA('datos') },
  ].filter(Boolean);
}

function itemsJuego() {
  const e = S.estado;
  if (!e) return sinEstado();
  const j = e.juego, sun = e.sunshine;
  return [
    { id: 'wow', icono: 'espada', titulo: 'World of Warcraft', sub: textoJuego(j), tono: j.wow ? 'ok' : '', accion: hojaWow },
    {
      id: 'reiniciar_wow', icono: 'refrescar', titulo: 'Reiniciar WoW',
      sub: j.wow ? 'Cierra el juego; Battle.net queda abierto' : 'WoW no está abierto',
      accion: j.wow ? confirmarReiniciarWow : null,
    },
    {
      id: 'battlenet', icono: 'play', titulo: 'Abrir Battle.net',
      sub: j.battlenet ? 'Ya está abierto' : 'Abre el launcher en la PC',
      accion: j.battlenet ? null : abrirBattlenet,
    },
    {
      id: 'terminar', icono: 'apagar', titulo: 'Terminar la sesión',
      sub: sun.ocupado ? 'Cierra todo y corta el stream' : 'No hay ninguna sesión abierta',
      accion: sun.ocupado ? confirmarTerminar : null,
    },
  ];
}

function itemsDatos() {
  const e = S.estado;
  if (!e) return sinEstado();
  const d = e.datos, c = d.config, via = viaCelu(d.celu);
  const total14 = d.dias.reduce((a, x) => a + x.bytes, 0);
  return [
    {
      id: 'ciclo', icono: 'senal', titulo: `Este ciclo · desde el ${fechaCorta(d.inicio)}`, sub: textoCiclo(d),
      tono: tonoDatos(d), pct: d.pct, accion: hojaHistorial,
    },
    { id: 'hoy', icono: 'calendario', titulo: 'Hoy', sub: tamano(d.hoy), accion: hojaHistorial },
    {
      id: 'sesion', icono: 'reloj', titulo: 'Sesión actual',
      sub: d.sesion == null ? 'No hay una sesión abierta' : `${tamano(d.sesion)} fuera de casa`,
    },
    { id: 'conexion', icono: via.icono, titulo: 'Conexión del celu', sub: via.texto, tono: via.tono, accion: hojaCelu },
    {
      id: 'limite', icono: 'escudo', titulo: 'Límite del ciclo',
      sub: c.limite_gb
        ? `${gb(c.limite_gb)} · ${c.accion === 'cortar' ? 'corta la sesión' : 'solo avisa'}`
        : 'Sin límite · tocá para poner uno',
      accion: hojaLimite,
    },
    {
      id: 'corte', icono: 'calendario', titulo: 'Día de corte',
      sub: `El ${c.dia_corte} de cada mes`, accion: hojaDiaCorte,
    },
    { id: 'historial', icono: 'barras', titulo: 'Últimos 14 días', sub: `${tamano(total14)} en total`, accion: hojaHistorial },
    { id: 'cero', icono: 'borrar', titulo: 'Poner el contador en cero', sub: 'Si tu compañía ya renovó el plan', accion: confirmarCero },
    { id: 'que', icono: 'info', titulo: '¿Qué se cuenta?', sub: 'Solo el stream fuera de casa', accion: hojaQueSeCuenta },
  ];
}

function itemsDispositivos() {
  const items = [];
  if (S.clientes == null) {
    items.push({ id: 'cargando', icono: 'reloj', titulo: 'Cargando dispositivos…', sub: '' });
  } else if (!Array.isArray(S.clientes)) {
    items.push({ id: 'err', icono: 'info', titulo: 'No pude leer la lista', sub: mensajeError(S.clientes.error), tono: 'error', accion: cargarClientes });
  } else {
    for (const c of S.clientes) {
      items.push({ id: `cli-${c.uuid}`, icono: 'celu', titulo: c.name || 'Sin nombre', sub: 'Emparejado con Sunshine', accion: () => hojaCliente(c) });
    }
  }
  items.push({ id: 'emparejar', icono: 'mas', titulo: 'Emparejar un dispositivo', sub: 'Con el PIN que muestra Moonlight', accion: hojaPin });
  return items;
}

function itemsSistema() {
  const e = S.estado;
  if (!e) return sinEstado();
  const pc = e.pc, g = pc.gpu;
  return [
    g
      ? {
        id: 'gpu', icono: 'chip', titulo: g.nombre,
        sub: `${NF0.format(g.temp)} °C · ${NF0.format(g.uso)}% de uso · ${tamanoBin(g.vram_usada)} de ${tamanoBin(g.vram_total)}`,
        tono: g.temp >= 83 ? 'error' : g.temp >= 75 ? 'alerta' : '',
      }
      : { id: 'gpu', icono: 'chip', titulo: 'Placa de video', sub: 'Sin datos' },
    { id: 'cpu', icono: 'pulso', titulo: 'Procesador', sub: pc.cpu == null ? 'Midiendo…' : `${pc.cpu}% de uso` },
    { id: 'ram', icono: 'memoria', titulo: 'Memoria', sub: `${tamanoBin(pc.ram_usada)} de ${tamanoBin(pc.ram_total)}` },
    { id: 'encendida', icono: 'reloj', titulo: 'Encendida hace', sub: pc.encendida ? duracion(pc.encendida) : '—' },
    { id: 'reiniciar_sunshine', icono: 'refrescar', titulo: 'Reiniciar Sunshine', sub: 'Si el stream no conecta', accion: confirmarReiniciarSunshine },
    { id: 'registros', icono: 'lista', titulo: 'Registros de Sunshine', sub: 'Últimos mensajes', accion: hojaRegistros },
    { id: 'probar_aviso', icono: 'campana', titulo: 'Probar un aviso', sub: 'Muestra una notificación en la PC', accion: probarAviso },
  ];
}

function itemsAjustes() {
  const tema = TEMAS.find(t => t.id === S.prefs.tema) || TEMAS[0];
  const version = EN_APP ? window.CloudGaming.version() : null;
  const panel = [
    { id: 'tema', icono: 'paleta', titulo: 'Color del fondo', sub: tema.nombre, accion: hojaTema },
    { id: 'ondas', icono: 'ondas', titulo: 'Ondas animadas', sub: S.prefs.ondas ? 'Activadas' : 'Desactivadas · ahorra batería', accion: () => alternar('ondas') },
    { id: 'sonido', icono: 'sonido', titulo: 'Sonidos y vibración', sub: S.prefs.sonido ? 'Activados' : 'Desactivados', accion: () => alternar('sonido') },
    {
      id: 'musica', icono: 'musica', titulo: 'Música de fondo', accion: hojaMusica,
      sub: S.prefs.musica === 'no' ? 'Apagada' : `${NOMBRES_MUSICA[S.prefs.musica]} · ambiental, nunca suena igual`,
    },
    EN_APP
      ? {
        id: 'app', icono: 'celu', titulo: 'App CloudGaming', accion: hojaInstalar,
        sub: version === VERSION_APP ? `Versión ${version}` : `Versión ${version} · actualizá a la ${VERSION_APP}`,
        tono: version === VERSION_APP ? '' : 'alerta',
      }
      : { id: 'app', icono: 'celu', titulo: 'App para Android', sub: 'Para jugar desde este menú', accion: hojaInstalar },
    CAMBIA_DE_PC ? { id: 'pc', icono: 'pc', titulo: 'Cambiar de PC', sub: 'Buscar las PC de tu red', accion: confirmarCambiarPc } : null,
    EN_APP ? null : {
      id: 'completa', icono: 'expandir', titulo: 'Pantalla completa',
      sub: document.fullscreenElement ? 'Activada · tocá para salir' : 'Oculta las barras del navegador', accion: pantallaCompleta,
    },
    S.porCuenta ? null : { id: 'salir', icono: 'salir', titulo: 'Cerrar sesión del panel', sub: 'Vas a tener que volver a entrar', accion: confirmarSalir },
    { id: 'acerca', icono: 'info', titulo: 'Panel CloudGaming', sub: 'Hecho a medida para tu PC', accion: hojaAcerca },
  ].filter(Boolean);
  const stream = itemsStream();
  if (stream.length) panel[0].grupo = 'Menú';
  return [...stream, ...panel];
}

// ------------------------------------------------------- ajustes del stream
// Los guarda la app en el celu (son los ajustes de Moonlight, manejados desde acá).

const CODECS = {
  auto: ['Automático', 'La app elige · recomendado'],
  forceh265: ['HEVC (H.265)', 'Mejor imagen con la misma calidad'],
  neverh265: ['H.264', 'El más compatible'],
};
const BUFERES = [
  [0, 'Mínimo', 'El de Moonlight original · se corta en wifi'],
  [40, '40 ms', 'Poca demora'],
  [80, '80 ms', 'Recomendado'],
  [120, '120 ms', 'Más estable'],
  [200, '200 ms', 'Para wifi muy inestable'],
];
const TECLADOS = {
  latam: 'Latinoamericano',
  es: 'Español (España)',
  us: 'Inglés (EE.UU.)',
};

function leerStream() {
  try {
    return JSON.parse(window.CloudGaming.ajustesStream());
  } catch {
    return null;
  }
}

const mbps = kbps => `${NF1.format(kbps / 1000)} Mbps`;
const gbHora = kbps => `${NF1.format(kbps * 3600 / 8 / 1e6)} GB`; // kbit/s → GB por hora
const textoRes = res => res.replace('x', ' × ');

function itemsStream() {
  const a = STREAM_PROPIO ? leerStream() : null;
  if (!a) return [];
  const bitrate = a.bitrate || a.bitrateAuto;
  const bufer = BUFERES.find(b => b[0] === a.bufferAudio);
  return [
    {
      id: 's-res', icono: 'tele', titulo: 'Resolución', grupo: 'Stream', accion: () => hojaResolucion(a),
      sub: a.resolucion === a.nativa ? `${textoRes(a.resolucion)} · nativa del celu` : textoRes(a.resolucion),
    },
    { id: 's-fps', icono: 'pulso', titulo: 'Cuadros por segundo', sub: `${a.fps} fps`, accion: () => hojaFps(a) },
    {
      id: 's-bitrate', icono: 'senal', titulo: 'Calidad de video', accion: () => hojaBitrate(a),
      sub: `${a.bitrate ? '' : 'Automática · '}${mbps(bitrate)} · ≈ ${gbHora(bitrate)} por hora`,
    },
    { id: 's-codec', icono: 'chip', titulo: 'Códec de video', sub: (CODECS[a.codec] || [a.codec])[0], accion: () => hojaCodec(a) },
    {
      id: 's-estirar', icono: 'expandir', titulo: 'Estirar la imagen', accion: () => alternarStream('estirar', a),
      sub: a.estirar ? 'Activado · llena toda la pantalla' : 'Desactivado · puede quedar con bandas',
    },
    {
      id: 's-bufer', icono: 'ondas', titulo: 'Búfer de audio', accion: () => hojaBufer(a),
      sub: bufer ? `${bufer[1]} · ${bufer[2]}` : `${a.bufferAudio} ms`,
    },
    {
      id: 's-audiopc', icono: 'sonido', titulo: 'Sonido también en la PC', accion: () => alternarStream('audioPc', a),
      sub: a.audioPc ? 'Activado · suena en los dos lados' : 'Desactivado · solo en el celu',
    },
    {
      id: 's-stats', icono: 'barras', titulo: 'Estadísticas en pantalla', accion: () => alternarStream('estadisticas', a),
      sub: a.estadisticas ? 'Activadas · fps, demora y pérdidas' : 'Desactivadas',
    },
    // a.modoMouse llega desde la app 2.4
    'modoMouse' in a ? {
      id: 's-mouse', icono: 'mando', titulo: 'Modo mouse (mantener Start)', accion: () => alternarStream('modoMouse', a),
      sub: a.modoMouse ? 'Activado · el stick mueve el mouse, A y B hacen clic' : 'Desactivado',
    } : null,
    // a.tecladoPc llega desde la app 3.6
    'tecladoPc' in a ? {
      id: 's-teclado', icono: 'teclado', titulo: 'Teclado de la PC', accion: () => hojaTeclado(a),
      sub: a.tecladoPc === 'auto' ? `Automático · ${tecladoInformado(a.tecladoInformado)}` : TECLADOS[a.tecladoPc],
    } : null,
  ].filter(Boolean);
}

function guardarStream(clave, valor) {
  if (!window.CloudGaming.guardarAjuste(clave, String(valor))) {
    avisar('No se pudo guardar', true);
    return false;
  }
  render();
  return true;
}

function alternarStream(clave, a) {
  if (guardarStream(clave, !a[clave])) sonar('ok');
}

/* opciones: [valor, texto, sub]; al elegir guarda y cierra. */
function hojaElegir(titulo, clave, nombreIcono, opciones, actual, nota) {
  abrirHoja(titulo, [
    nota || null,
    ...opciones.map(([valor, texto, sub]) => opcion(texto, nombreIcono, () => {
      if (guardarStream(clave, valor)) {
        cerrarHoja();
        avisar(`${titulo}: ${texto}`);
      }
    }, { sub, activa: String(valor) === String(actual) })),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

function hojaResolucion(a) {
  const opciones = [
    ['1280x720', '1280 × 720', 'HD · la más liviana'],
    ['1920x1080', '1920 × 1080', 'Full HD'],
    ['2560x1440', '2560 × 1440', '2K'],
    [a.nativa, textoRes(a.nativa), 'Nativa del celu · la más nítida, sin bandas'],
  ];
  if (!opciones.some(o => o[0] === a.resolucion)) opciones.push([a.resolucion, textoRes(a.resolucion), 'La que tenías']);
  hojaElegir('Resolución', 'resolucion', 'tele', opciones, a.resolucion,
    parrafo('Más resolución se ve mejor pero pide más ', ['calidad de video'], ' (y gasta más datos fuera de casa).'));
}

function hojaFps(a) {
  const opciones = [30, 60, 90, 120].filter(f => f <= a.maxFps + 1)
    .map(f => [String(f), `${f} fps`, f === 60 ? 'Recomendado' : f > 60 ? 'Más fluido, pide más calidad de video' : 'Ahorra datos y batería']);
  hojaElegir('Cuadros por segundo', 'fps', 'pulso', opciones, a.fps);
}

function hojaBitrate(a) {
  const opciones = [['auto', `Automática · ${mbps(a.bitrateAuto)}`, `Según resolución y fps · ≈ ${gbHora(a.bitrateAuto)} por hora`]];
  for (const m of [5, 10, 15, 20, 30, 40, 60]) opciones.push([String(m * 1000), `${m} Mbps`, `≈ ${gbHora(m * 1000)} por hora`]);
  hojaElegir('Calidad de video', 'bitrate', 'senal', opciones, a.bitrate ? String(a.bitrate) : 'auto',
    parrafo('Más Mbps = imagen más limpia, pero ', ['gasta más datos'], ' fuera de casa y necesita mejor señal.'));
}

function hojaCodec(a) {
  hojaElegir('Códec de video', 'codec', 'chip',
    Object.entries(CODECS).map(([valor, [texto, sub]]) => [valor, texto, sub]), a.codec);
}

function hojaBufer(a) {
  hojaElegir('Búfer de audio', 'bufferAudio', 'ondas', BUFERES, a.bufferAudio,
    parrafo('Guarda un poco de sonido de reserva para que el wifi no lo corte. Más búfer = ',
      ['menos cortes'], ', pero el sonido llega un poquito después que la imagen.'));
}

/* Lo que informó la PC al conectar (vacío si todavía no conectó con la app 3.6 o el host no lo dice). */
function tecladoInformado(nombre) {
  if (!nombre) return 'se detecta al jugar';
  if (TECLADOS[nombre]) return TECLADOS[nombre];
  return nombre.startsWith('windows') ? 'otro de Windows' : `otro (${nombre}), sin traducir`;
}

function hojaTeclado(a) {
  hojaElegir('Teclado de la PC', 'tecladoPc', 'teclado', [
    ['auto', 'Automático', `El que usa la PC · ${tecladoInformado(a.tecladoInformado)}`],
    ...Object.entries(TECLADOS).map(([valor, texto]) => [valor, texto, '']),
  ], a.tecladoPc,
  parrafo('Para que los símbolos y las tildes del teclado del celu salgan bien en la PC (', ['@ ¿ ñ á'],
    '), la app tiene que saber qué teclado usa la PC. Cambialo solo si salen mal.'));
}

const CATEGORIAS = [
  { id: 'jugar', nombre: 'Jugar', icono: 'mando', items: itemsJugar },
  { id: 'estado', nombre: 'Estado', icono: 'pulso', items: itemsEstado },
  { id: 'juego', nombre: 'WoW', icono: 'espada', items: itemsJuego },
  { id: 'datos', nombre: 'Datos', icono: 'senal', items: itemsDatos },
  { id: 'dispositivos', nombre: 'Dispositivos', icono: 'celu', items: itemsDispositivos, alEntrar: cargarClientes },
  { id: 'sistema', nombre: 'Sistema', icono: 'chip', items: itemsSistema },
  { id: 'ajustes', nombre: 'Ajustes', icono: 'ajustes', items: itemsAjustes },
];

// ------------------------------------------------------------------- dibujo

const catsEl = $('#categorias');
const itemsEl = $('#items');
let catNodos = [];
const itemNodos = new Map();
let itemsDeCat = null;

function construirCategorias() {
  catsEl.textContent = '';
  catNodos = CATEGORIAS.map((c, i) => {
    const el = crear('div', 'cat');
    el.dataset.i = i;
    const ico = crear('span', 'ico');
    ico.innerHTML = icono(c.icono);
    el.append(ico, crear('span', 'nombre', c.nombre));
    catsEl.append(el);
    return el;
  });
}

function medidas() {
  const w = innerWidth, h = innerHeight, angosto = w < 560;
  const slot = angosto ? 92 : 120;
  const catX = Math.round(limitar(w * 0.22, slot * 0.95, 300));
  const catY = Math.round(limitar(h * (h > w ? 0.17 : 0.2), 84, 220));
  const catH = angosto ? 92 : 100;
  const itemH = angosto ? 60 : 66;
  return {
    w, slot, catX, catY, itemH,
    itemSel: itemH + 22,
    itemX: Math.round(catX + slot / 2 - 20), // centro del ícono bajo el de la categoría
    y0: catY + catH + 6,
    tope: 64,
  };
}

function itemsActuales() {
  return CATEGORIAS[S.cat].items();
}

function indiceSel(cantidad) {
  const id = CATEGORIAS[S.cat].id;
  S.sel[id] = limitar(S.sel[id] ?? 0, 0, Math.max(cantidad - 1, 0));
  return S.sel[id];
}

function nodoItem() {
  const el = crear('li', 'item');
  el.setAttribute('role', 'option');
  const txt = crear('span', 'txt');
  txt.append(crear('span', 'tit'), crear('span', 'sub'));
  el.append(crear('span', 'ico'), txt);
  el.addEventListener('animationend', () => el.classList.remove('entrando'));
  return el;
}

function pintarItem(el, it) {
  const ico = el.firstChild, txt = el.lastChild;
  const tit = txt.querySelector('.tit'), sub = txt.querySelector('.sub');
  let grupo = txt.querySelector('.grupo');
  if (it.grupo) {
    if (!grupo) {
      grupo = crear('span', 'grupo');
      txt.prepend(grupo);
    }
    if (grupo.textContent !== it.grupo) grupo.textContent = it.grupo;
  } else if (grupo) {
    grupo.remove();
  }
  if (el.dataset.icono !== it.icono) {
    ico.innerHTML = icono(it.icono);
    el.dataset.icono = it.icono;
  }
  if (tit.textContent !== it.titulo) tit.textContent = it.titulo;
  const texto = it.sub || '';
  if (sub.textContent !== texto) sub.textContent = texto;
  el.classList.remove('ok', 'alerta', 'error');
  if (it.tono) el.classList.add(it.tono);
  let barra = txt.querySelector('.barrita');
  if (it.pct != null) {
    if (!barra) {
      barra = crear('span', 'barrita');
      barra.append(crear('i'));
      txt.append(barra);
    }
    barra.firstChild.style.setProperty('--pct', `${limitar(it.pct, 0, 100)}%`);
  } else if (barra) {
    barra.remove();
  }
}

function render(cambioCat = false) {
  const m = medidas();
  catNodos.forEach((el, i) => {
    el.style.width = `${m.slot}px`;
    el.style.transform = `translate(${m.catX + (i - S.cat) * m.slot}px, ${m.catY}px)`;
    el.classList.toggle('sel', i === S.cat);
  });

  const items = itemsActuales();
  const cat = CATEGORIAS[S.cat].id;
  const sel = indiceSel(items.length);
  if (cambioCat || itemsDeCat !== cat) {
    itemsEl.textContent = '';
    itemNodos.clear();
    itemsDeCat = cat;
  }
  const vivos = new Set();
  items.forEach((it, j) => {
    let el = itemNodos.get(it.id);
    if (!el) {
      el = nodoItem();
      el.classList.add('entrando');
      if (cambioCat) el.style.animationDelay = `${Math.min(j, 6) * 28}ms`;
      itemNodos.set(it.id, el);
      itemsEl.append(el);
    }
    pintarItem(el, it);
    vivos.add(it.id);
    let y;
    if (j === sel) y = m.y0;
    else if (j > sel) y = m.y0 + m.itemSel + (j - sel - 1) * m.itemH;
    else y = m.catY - m.itemH * (sel - j) - 6;
    el.style.transform = `translate(${m.itemX}px, ${y}px)`;
    el.style.width = `${m.w - m.itemX - 16}px`;
    el.dataset.j = j;
    el.classList.toggle('sel', j === sel);
    el.classList.toggle('arriba', j < sel);
    el.classList.toggle('oculto', j < sel && y < m.tope);
    el.setAttribute('aria-selected', j === sel);
  });
  for (const [clave, el] of itemNodos) {
    if (!vivos.has(clave)) {
      el.remove();
      itemNodos.delete(clave);
    }
  }
}

function chip(tono, nombreIcono, texto) {
  const c = crear('span', `chip${tono ? ` ${tono}` : ''}`);
  if (nombreIcono) c.innerHTML = icono(nombreIcono);
  else c.append(crear('span', 'punto'));
  c.append(crear('span', null, texto));
  return c;
}

function pintarIndicadores() {
  const cont = $('#indicadores');
  cont.textContent = '';
  if ($('#menu').hidden) return;
  if (S.errorEstado) {
    if (!sinRed()) cont.append(chip('error', null, 'Sin conexión con la PC'));
    else if (RED_PROPIA) cont.append(chip('alerta', null, 'Conectando…'));
    else cont.append(chip('error', null, 'Tailscale apagado'));
    return;
  }
  const e = S.estado;
  if (!e) return;
  const sun = e.sunshine, d = e.datos, via = viaCelu(d.celu);
  cont.append(chip(!sun.activo ? 'error' : sun.ocupado ? 'ok' : '', null,
    !sun.activo ? 'Sunshine apagado' : sun.ocupado ? 'En vivo' : 'Listo'));
  cont.append(chip(via.tono === 'alerta' || via.tono === 'error' ? via.tono : '', via.icono, via.corto));
  if (d.limite) cont.append(chip(tonoDatos(d), 'senal', `${tamano(d.ciclo)} / ${gb(d.config.limite_gb)}`));
  else if (d.ciclo) cont.append(chip('', 'senal', tamano(d.ciclo)));
}

function pintarReloj() {
  const d = new Date();
  const dia = d.toLocaleDateString('es-AR', { weekday: 'short' }).replace('.', '');
  $('#fecha').textContent = `${dia} ${d.getDate()}/${d.getMonth() + 1}`;
  $('#hora').textContent = d.toLocaleTimeString('es-AR', { hour: '2-digit', minute: '2-digit', hourCycle: 'h23' });
}

// -------------------------------------------------------------- navegación

function irACat(i) {
  if (i === S.cat || i < 0 || i >= CATEGORIAS.length) return;
  S.cat = i;
  history.replaceState(null, '', `#${CATEGORIAS[i].id}`);
  sonar('cat');
  render(true);
  CATEGORIAS[i].alEntrar?.();
  contarPaso();
}

const moverCat = d => irACat(limitar(S.cat + d, 0, CATEGORIAS.length - 1));
const irA = id => irACat(CATEGORIAS.findIndex(c => c.id === id));

function moverItem(d) {
  const items = itemsActuales();
  const id = CATEGORIAS[S.cat].id;
  const actual = indiceSel(items.length);
  const nuevo = limitar(actual + d, 0, items.length - 1);
  if (nuevo === actual) return;
  S.sel[id] = nuevo;
  sonar('item');
  render();
  contarPaso();
}

function activar() {
  const items = itemsActuales();
  const it = items[indiceSel(items.length)];
  if (it?.accion) it.accion();
  else sonar('nada');
}

function contarPaso() {
  if (S.prefs.ayudaVista) return;
  if (++S.pasos >= 4) {
    S.prefs.ayudaVista = true;
    guardarPrefs();
    $('#ayuda').classList.add('vista');
  }
}

function tocar(objetivo) {
  const cat = objetivo.closest('.cat');
  if (cat) {
    irACat(+cat.dataset.i);
    return;
  }
  const item = objetivo.closest('.item');
  if (!item) return;
  const j = +item.dataset.j;
  const id = CATEGORIAS[S.cat].id;
  if (j === S.sel[id]) {
    activar();
  } else {
    S.sel[id] = j;
    sonar('item');
    render();
  }
}

let toque = null;

$('#menu').addEventListener('pointerdown', e => {
  if (e.button > 0) return;
  S.entrada = 'toque';
  toque = { x: e.clientX, y: e.clientY, id: e.pointerId, eje: null, pasos: 0 };
});

addEventListener('pointermove', e => {
  if (!toque || e.pointerId !== toque.id) return;
  const dx = e.clientX - toque.x, dy = e.clientY - toque.y;
  if (!toque.eje && Math.hypot(dx, dy) > 12) toque.eje = Math.abs(dx) > Math.abs(dy) ? 'x' : 'y';
  if (toque.eje === 'y') {
    // Arrastrar para arriba baja por la lista, como al deslizar una página.
    const pasos = Math.trunc(-dy / (medidas().itemH * 0.75));
    if (pasos !== toque.pasos) {
      moverItem(pasos - toque.pasos);
      toque.pasos = pasos;
    }
  }
});

addEventListener('pointerup', e => {
  if (!toque || e.pointerId !== toque.id) return;
  const t = toque;
  toque = null;
  const dx = e.clientX - t.x;
  if (t.eje === 'x') {
    if (Math.abs(dx) > 36) moverCat(dx < 0 ? 1 : -1);
    return;
  }
  if (t.eje === 'y') return;
  tocar(e.target);
});

addEventListener('pointercancel', () => { toque = null; });

let ultimaRueda = 0;
$('#menu').addEventListener('wheel', e => {
  const ahora = performance.now();
  if (ahora - ultimaRueda < 120) return;
  ultimaRueda = ahora;
  if (Math.abs(e.deltaX) > Math.abs(e.deltaY)) moverCat(Math.sign(e.deltaX));
  else moverItem(Math.sign(e.deltaY));
}, { passive: true });

addEventListener('keydown', e => {
  if (!$('#entrada').hidden) return;
  S.entrada = 'teclas';
  if (!hoja.hidden) {
    teclaHoja(e);
    return;
  }
  const acciones = {
    ArrowLeft: () => moverCat(-1), ArrowRight: () => moverCat(1),
    ArrowUp: () => moverItem(-1), ArrowDown: () => moverItem(1),
    Enter: activar, ' ': activar,
  };
  const accion = acciones[e.key];
  if (!accion) return;
  e.preventDefault();
  accion();
});

let redibujo = 0;
addEventListener('resize', () => {
  cancelAnimationFrame(redibujo);
  redibujo = requestAnimationFrame(() => render());
});

document.addEventListener('fullscreenchange', () => render());

// ----------------------------------------------------------------- joystick
// El navegador solo deja leer joysticks en páginas seguras (HTTPS o la propia PC).

const mando = { activo: false, repetir: {} };

function accionMando(boton) {
  S.entrada = 'teclas';
  if (!$('#entrada').hidden) return;
  if (!hoja.hidden) {
    if (boton === 'arriba' || boton === 'izq') moverFocoHoja(-1);
    else if (boton === 'abajo' || boton === 'der') moverFocoHoja(1);
    else if (boton === 'a') {
      if (hoja.contains(document.activeElement) && document.activeElement !== document.body) document.activeElement.click();
      else moverFocoHoja(1);
    } else if (boton === 'b') cerrarHoja();
    return;
  }
  ({ arriba: () => moverItem(-1), abajo: () => moverItem(1), izq: () => moverCat(-1), der: () => moverCat(1), a: activar })[boton]?.();
}

function leerMando() {
  const pad = [...(navigator.getGamepads?.() || [])].find(Boolean);
  if (!pad) {
    mando.activo = false;
    return;
  }
  const b = i => pad.buttons[i]?.pressed;
  const ax = pad.axes[0] || 0, ay = pad.axes[1] || 0;
  const estado = {
    arriba: b(12) || ay < -0.6, abajo: b(13) || ay > 0.6,
    izq: b(14) || ax < -0.6, der: b(15) || ax > 0.6, a: b(0), b: b(1),
  };
  const ahora = performance.now();
  for (const [boton, apretado] of Object.entries(estado)) {
    if (!apretado) {
      delete mando.repetir[boton];
      continue;
    }
    const proximo = mando.repetir[boton];
    if (proximo == null) {
      mando.repetir[boton] = ahora + 380;
      accionMando(boton);
    } else if (ahora >= proximo && boton !== 'a' && boton !== 'b') {
      mando.repetir[boton] = ahora + 110;
      accionMando(boton);
    }
  }
  requestAnimationFrame(leerMando);
}

addEventListener('gamepadconnected', () => {
  if (EN_APP || mando.activo) return; // en la app el joystick llega como teclas
  mando.activo = true;
  requestAnimationFrame(leerMando);
});

// --------------------------------------------------------------------- hoja

const hoja = $('#hoja');
const hojaTitulo = $('#hoja-titulo');
const hojaCuerpo = $('.hoja-cuerpo');
let hojaTimer = null;

function abrirHoja(titulo, contenido) {
  clearInterval(hojaTimer);
  hojaTimer = null;
  hojaTitulo.textContent = titulo;
  hojaCuerpo.textContent = '';
  for (const nodo of [].concat(contenido)) if (nodo) hojaCuerpo.append(nodo);
  hojaCuerpo.parentElement.scrollTop = 0;
  if (hoja.hidden) {
    hoja.hidden = false;
    requestAnimationFrame(() => requestAnimationFrame(() => hoja.classList.add('abierta')));
  }
  sonar('abrir');
  if (S.entrada !== 'toque') {
    // Con joystick o teclado arranca con la primera opción marcada.
    setTimeout(() => focoHoja().find(el => el.tagName === 'BUTTON')?.focus(), 60);
  }
}

function cerrarHoja() {
  if (hoja.hidden) return;
  clearInterval(hojaTimer);
  hojaTimer = null;
  hoja.classList.remove('abierta');
  if (hoja.contains(document.activeElement)) document.activeElement.blur();
  setTimeout(() => { if (!hoja.classList.contains('abierta')) hoja.hidden = true; }, 330);
  sonar('atras');
}

hoja.querySelector('.hoja-fondo').addEventListener('click', cerrarHoja);

const focoHoja = () => [...hojaCuerpo.querySelectorAll('button:not(:disabled), input')];
const esCampo = el => el?.tagName === 'INPUT';

function moverFocoHoja(d) {
  const lista = focoHoja();
  if (!lista.length) return;
  const i = limitar(lista.indexOf(document.activeElement) + d, 0, lista.length - 1);
  lista[i].focus();
  lista[i].scrollIntoView({ block: 'nearest' });
  sonar('item');
}

function teclaHoja(e) {
  if (e.key === 'Escape' || (e.key === 'Backspace' && !esCampo(e.target))) {
    e.preventDefault();
    cerrarHoja();
  } else if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    if (esCampo(e.target) && e.target.type === 'number') return;
    e.preventDefault();
    moverFocoHoja(e.key === 'ArrowDown' ? 1 : -1);
  }
}

function opcion(texto, nombreIcono, accion, { sub, peligro, activa, deshabilitada } = {}) {
  const b = crear('button', `opcion${peligro ? ' peligro' : ''}${activa ? ' activa' : ''}`);
  b.type = 'button';
  b.innerHTML = icono(nombreIcono);
  const t = crear('span');
  t.append(document.createTextNode(texto));
  if (sub) t.append(crear('small', null, sub));
  b.append(t);
  b.disabled = !!deshabilitada;
  b.addEventListener('click', () => {
    sonar('ok');
    accion();
  });
  return b;
}

/* parrafo('Texto normal ', ['en negrita'], ' y más texto') */
function parrafo(...partes) {
  const p = crear('p');
  for (const parte of partes) {
    if (Array.isArray(parte)) p.append(crear('strong', null, parte.join('')));
    else p.append(document.createTextNode(parte));
  }
  return p;
}

const seccion = texto => crear('p', 'seccion', texto);

function fichas(valores, actual, alElegir) {
  const cont = crear('div', 'fichas');
  const botones = valores.map(([valor, texto]) => {
    const b = crear('button', 'ficha', texto);
    b.type = 'button';
    b.dataset.valor = valor;
    b.addEventListener('click', () => {
      alElegir(valor);
      marcar(valor);
      sonar('item');
    });
    cont.append(b);
    return b;
  });
  function marcar(valor) {
    botones.forEach(b => b.classList.toggle('activa', b.dataset.valor === String(valor)));
  }
  marcar(actual);
  cont.marcar = marcar;
  return cont;
}

function confirmar(titulo, texto, textoSi, nombreIcono, accion) {
  abrirHoja(titulo, [
    parrafo(texto),
    opcion(textoSi, nombreIcono, () => {
      cerrarHoja();
      accion();
    }, { peligro: true }),
    opcion('Cancelar', 'cerrar', cerrarHoja),
  ]);
}

// ------------------------------------------------------------------ avisos

let avisoTimer = 0;

function avisar(texto, error = false) {
  const el = $('#aviso');
  el.textContent = texto;
  el.classList.toggle('error', error);
  el.classList.add('visible');
  clearTimeout(avisoTimer);
  avisoTimer = setTimeout(() => el.classList.remove('visible'), error ? 4200 : 2800);
}

const avisarError = e => {
  if (e?.codigo !== 'sesion') avisar(mensajeError(e), true);
  sonar('nada');
};

// ----------------------------------------------------------------- acciones

async function hacer(accion, mensajeOk) {
  try {
    const r = await api('accion', { accion });
    avisar(typeof mensajeOk === 'function' ? mensajeOk(r) : mensajeOk);
    setTimeout(actualizar, 800);
    return r;
  } catch (e) {
    avisarError(e);
    return null;
  }
}

const confirmarTerminar = () => confirmar(
  'Terminar la sesión',
  'Se cierra la app en Sunshine (con WoW y Battle.net) y se corta el stream.',
  'Terminar la sesión', 'apagar',
  () => hacer('cerrar_app', r => (r.ok ? 'Sesión terminada' : 'Sunshine no cerró la app')),
);

const confirmarReiniciarWow = () => confirmar(
  'Reiniciar WoW',
  'Se cierra el juego a la fuerza. Battle.net queda abierto: después elegilo en «Jugar» para volver a entrar.',
  'Cerrar WoW', 'refrescar',
  () => hacer('cerrar_wow', r => (r.estaba ? 'WoW cerrado. Tocá «Jugar» en Battle.net' : 'WoW ya estaba cerrado')),
);

const abrirBattlenet = () => hacer('abrir_battlenet', r => (r.abierto ? 'Abriendo Battle.net…' : 'Battle.net ya estaba abierto'));

const confirmarReiniciarSunshine = () => confirmar(
  'Reiniciar Sunshine',
  'Si hay un stream abierto se corta. Sunshine tarda unos 10 segundos en volver.',
  'Reiniciar', 'refrescar',
  () => hacer('reiniciar_sunshine', 'Reiniciando Sunshine…'),
);

function jugar(app) {
  const m = S.estado?.moonlight;
  if (!m?.uuid) {
    avisar('No pude leer los datos de la PC', true);
    return;
  }
  if (EN_APP) {
    Musica.alJugar();
    window.CloudGaming.jugar(m.uuid, m.pc || '', app);
    avisar(`Abriendo ${app}…`);
    return;
  }
  if (EN_ANDROID) {
    // Chrome no deja abrir Moonlight directo: se lo pide a la app, y si no está, ofrece instalarla.
    const consulta = Object.entries({ uuid: m.uuid, pc: m.pc || '', app })
      .map(([clave, valor]) => `${clave}=${encodeURIComponent(valor)}`).join('&');
    const respaldo = encodeURIComponent(`${location.origin}${location.pathname}?instalar=1#jugar`);
    Musica.alJugar();
    location.href = `intent://jugar?${consulta}#Intent;scheme=cloudgaming;package=${PAQUETE_APP};S.browser_fallback_url=${respaldo};end`;
    return;
  }
  avisar('Para jugar, abrí este menú en el celu', true);
}

const probarAviso = () => hacer('probar_aviso', 'Aviso enviado a la pantalla de la PC');

const confirmarSalir = () => confirmar(
  'Cerrar sesión', 'Para volver a entrar vas a necesitar el usuario y la contraseña de Sunshine.',
  'Cerrar sesión', 'salir',
  async () => {
    try { await api('logout', {}); } catch { /* igual salimos */ }
    mostrarEntrada();
  },
);

const confirmarCambiarPc = () => confirmar(
  'Cambiar de PC',
  'La app va a buscar las PC con CloudGaming en tu red para que elijas con cuál jugar.',
  'Buscar', 'pc',
  () => window.CloudGaming.cambiarPc(),
);

const confirmarCero = () => confirmar(
  'Poner el contador en cero',
  'El consumo de este ciclo vuelve a cero. El historial por día no se borra.',
  'Poner en cero', 'borrar',
  async () => {
    try {
      S.estado.datos = await api('datos/cero', {});
      avisar('Contador en cero');
      render();
    } catch (e) {
      avisarError(e);
    }
  },
);

async function cargarClientes() {
  try {
    S.clientes = (await api('clientes')).clientes;
  } catch (e) {
    S.clientes = { error: e.codigo };
  }
  if (CATEGORIAS[S.cat].id === 'dispositivos') render();
}

function alternar(pref) {
  S.prefs[pref] = !S.prefs[pref];
  guardarPrefs();
  if (pref === 'ondas') Ondas.actualizar();
  sonar('ok');
  render();
}

function pantallaCompleta() {
  if (document.fullscreenElement) {
    document.exitFullscreen?.();
    return;
  }
  const pedido = document.documentElement.requestFullscreen?.({ navigationUI: 'hide' });
  if (!pedido) avisar('Este navegador no tiene pantalla completa', true);
  else pedido.catch(() => avisar('El navegador no lo permitió', true));
}

// ------------------------------------------------------------------- hojas

function hojaSunshine() {
  const sun = S.estado.sunshine;
  abrirHoja('Sunshine', [
    parrafo('Estado: ', [!sun.activo ? 'apagado' : sun.ocupado ? 'transmitiendo' : 'listo para jugar']),
    sun.ocupado ? parrafo('App: ', [sun.app || 'desconocida'], sun.desde ? ` · desde las ${horaCorta(sun.desde)}` : '') : null,
    sun.ocupado ? opcion('Terminar la sesión', 'apagar', confirmarTerminar, { peligro: true, sub: 'Cierra la app y corta el stream' }) : null,
    opcion('Reiniciar Sunshine', 'refrescar', confirmarReiniciarSunshine, { sub: 'Si el stream no conecta' }),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

function hojaCelu() {
  const celu = S.estado.datos.celu, via = viaCelu(celu);
  const explicacion = {
    casa: 'El celu y la PC se hablan por la red de tu casa: no gasta datos móviles.',
    datos: 'El celu está fuera de casa y se conecta directo a la PC. Esto sí gasta datos (o el wifi de donde estés).',
    rebote: 'No se pudo armar la conexión directa y Tailscale pasa por un servidor intermedio: anda, pero con más demora. Si pasa seguido, se arregla abriendo el puerto UDP 41641 en el router.',
  }[celu?.via] || 'Cuando el stream se conecte por Tailscale vas a ver acá por dónde va.';
  abrirHoja(celu?.nombre || 'Celu', [
    parrafo('Conexión: ', [via.texto]),
    celu?.ip ? parrafo('IP de Tailscale: ', [celu.ip]) : null,
    celu?.direccion ? parrafo('Dirección: ', [celu.direccion]) : null,
    parrafo(explicacion),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

function hojaWow() {
  const j = S.estado.juego;
  abrirHoja('World of Warcraft', [
    parrafo('Juego: ', [j.wow ? 'abierto' : 'cerrado'], ' · Battle.net: ', [j.battlenet ? 'abierto' : 'cerrado']),
    j.wow ? opcion('Cerrar WoW', 'cerrar', confirmarReiniciarWow, { peligro: true, sub: 'Para reiniciarlo si se colgó' }) : null,
    j.battlenet ? null : opcion('Abrir Battle.net', 'play', () => { cerrarHoja(); abrirBattlenet(); }, { sub: 'Se abre en la pantalla de la PC' }),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

function hojaHistorial() {
  const d = S.estado.datos;
  const maximo = Math.max(...d.dias.map(x => x.bytes), 1);
  const grafico = crear('div', 'grafico');
  d.dias.forEach((x, i) => {
    const col = crear('div', `col${i === d.dias.length - 1 ? ' hoy' : ''}`);
    const barra = crear('div', 'bar');
    barra.style.setProperty('--alto', `${(x.bytes / maximo * 100).toFixed(1)}%`);
    barra.title = `${fechaCorta(x.fecha)}: ${tamano(x.bytes)}`;
    col.append(barra, crear('span', 'dia', String(+x.fecha.slice(8))));
    grafico.append(col);
  });
  const mayor = d.dias.reduce((a, x) => (x.bytes > a.bytes ? x : a), d.dias[0]);
  abrirHoja('Consumo fuera de casa', [
    seccion('Últimos 14 días'),
    grafico,
    parrafo('Este ciclo: ', [tamano(d.ciclo)], ` · se renueva el ${fechaCorta(d.fin)}`),
    parrafo('Hoy: ', [tamano(d.hoy)]),
    mayor.bytes ? parrafo('Día con más consumo: ', [`${fechaCorta(mayor.fecha)} (${tamano(mayor.bytes)})`]) : null,
    parrafo('Ciclo anterior: ', [tamano(d.anterior)]),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

async function guardarDatos(cambios, mensaje) {
  try {
    S.estado.datos = await api('datos/config', cambios);
    cerrarHoja();
    avisar(mensaje);
    render();
    pintarIndicadores();
  } catch (e) {
    avisarError(e);
  }
}

function hojaLimite() {
  const c = { ...S.estado.datos.config };
  const otro = crear('input');
  otro.type = 'number';
  otro.inputMode = 'decimal';
  otro.min = '0';
  otro.step = '0.5';
  otro.placeholder = 'Otro valor en GB';
  otro.value = c.limite_gb || '';
  const fichasGb = fichas([0, 1, 2, 3, 5, 10, 20].map(v => [v, v ? `${v} GB` : 'Sin límite']), c.limite_gb, v => {
    c.limite_gb = v;
    otro.value = v || '';
  });
  otro.addEventListener('input', () => {
    c.limite_gb = parseFloat(otro.value.replace(',', '.')) || 0;
    fichasGb.marcar(c.limite_gb);
  });
  const avisar_ = opcion('Solo avisar', 'campana', () => elegirAccion('avisar'), { sub: 'Te muestra un aviso en la pantalla del juego', activa: c.accion === 'avisar' });
  const cortar = opcion('Cortar la sesión', 'apagar', () => elegirAccion('cortar'), { sub: 'Avisa y corta el stream si seguís con datos', activa: c.accion === 'cortar' });
  function elegirAccion(accion) {
    c.accion = accion;
    avisar_.classList.toggle('activa', accion === 'avisar');
    cortar.classList.toggle('activa', accion === 'cortar');
  }
  abrirHoja('Límite del ciclo', [
    parrafo('Cuánto querés gastar como máximo jugando fuera de casa en cada ciclo de tu plan.'),
    seccion('Límite'),
    fichasGb,
    otro,
    seccion('Avisarme al llegar al'),
    fichas([50, 70, 80, 90].map(v => [v, `${v}%`]), c.aviso_pct, v => { c.aviso_pct = v; }),
    seccion('Al llegar al límite'),
    avisar_,
    cortar,
    opcion('Guardar', 'escudo', () => guardarDatos(c, 'Límite guardado')),
    opcion('Cancelar', 'cerrar', cerrarHoja),
  ]);
}

function hojaDiaCorte() {
  let dia = S.estado.datos.config.dia_corte;
  const salida = crear('output', null, String(dia));
  const cambiar = d => {
    dia = limitar(dia + d, 1, 28);
    salida.textContent = String(dia);
    sonar('item');
  };
  const menos = crear('button', 'ficha', '−');
  const mas = crear('button', 'ficha', '+');
  menos.type = mas.type = 'button';
  menos.addEventListener('click', () => cambiar(-1));
  mas.addEventListener('click', () => cambiar(1));
  const paso = crear('div', 'paso');
  paso.append(menos, salida, mas);
  abrirHoja('Día de corte', [
    parrafo('El día del mes en que tu compañía renueva los datos. Ahí el contador vuelve a cero solo.'),
    paso,
    opcion('Guardar', 'calendario', () => guardarDatos({ dia_corte: dia }, `Día de corte: el ${dia}`)),
    opcion('Cancelar', 'cerrar', cerrarHoja),
  ]);
}

function hojaQueSeCuenta() {
  abrirHoja('¿Qué se cuenta?', [
    parrafo('Cuento lo que la PC y el celu se mandan por ', ['Tailscale'], ' cuando el celu está ', ['fuera de casa'], ': con datos móviles o en otro wifi.'),
    parrafo('En el wifi de casa no se cuenta, porque no gasta datos.'),
    parrafo('Es una cuenta ', ['aproximada'], ': tu compañía puede marcar un poco más. Lo que usan otras apps del celu (WhatsApp, YouTube…) no entra acá.'),
    parrafo('Si te pasás del límite, según lo que elijas, aparece un aviso en la pantalla del juego o se corta la sesión.'),
    parrafo('El contador corre en la PC: sigue midiendo aunque cierres esta página.'),
    opcion('Entendido', 'cerrar', cerrarHoja),
  ]);
}

function hojaCliente(c) {
  abrirHoja(c.name || 'Dispositivo', [
    parrafo('Este dispositivo puede conectarse a la PC con Moonlight.'),
    opcion('Desemparejar', 'desconectar', () => confirmar(
      'Desemparejar', `«${c.name || 'Este dispositivo'}» no va a poder conectarse más hasta que lo vuelvas a emparejar con el PIN.`,
      'Desemparejar', 'desconectar',
      async () => {
        try {
          await api('clientes/desemparejar', { uuid: c.uuid });
          avisar('Desemparejado');
          cargarClientes();
        } catch (e) {
          avisarError(e);
        }
      },
    ), { peligro: true }),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

function hojaPin() {
  let elegido = null;
  let ultimaLista = '';
  const lista = crear('div', 'hoja-cuerpo');
  const pin = crear('input');
  pin.inputMode = 'numeric';
  pin.maxLength = 4;
  pin.placeholder = 'PIN de 4 números';
  pin.autocomplete = 'one-time-code';
  const nombre = crear('input');
  nombre.placeholder = 'Nombre (por ejemplo: Tablet)';
  nombre.maxLength = 40;

  function pintar(pendientes) {
    lista.textContent = '';
    if (!pendientes.length) {
      elegido = null;
      lista.append(parrafo('Todavía no hay pedidos. En Moonlight, agregá la PC y tocá su ícono: te va a mostrar un PIN.'));
      return;
    }
    if (!pendientes.some(p => p.id === elegido)) elegido = pendientes.length === 1 ? pendientes[0].id : null;
    for (const p of pendientes) {
      lista.append(opcion(p.name || 'Dispositivo desconocido', 'celu', () => {
        elegido = p.id;
        if (!nombre.value) nombre.value = p.name || '';
        pintar(pendientes);
      }, { sub: p.address || 'Dirección desconocida', activa: p.id === elegido }));
    }
  }

  async function cargar() {
    try {
      const { pendientes } = await api('pin');
      const firma = JSON.stringify(pendientes);
      if (firma !== ultimaLista) {
        ultimaLista = firma;
        pintar(pendientes);
      }
    } catch (e) {
      ultimaLista = '';
      lista.textContent = '';
      lista.append(parrafo(mensajeError(e)));
    }
  }

  async function enviar() {
    if (!elegido) return avisar('Elegí el dispositivo que está pidiendo el PIN', true);
    if (!/^\d{4}$/.test(pin.value)) return avisar('El PIN tiene 4 números', true);
    try {
      const r = await api('pin', { pairing_id: elegido, pin: pin.value, name: nombre.value.trim() });
      if (!r.ok) return avisar('El PIN no coincide', true);
      cerrarHoja();
      avisar('¡Listo! Seguí en Moonlight');
      cargarClientes();
    } catch (e) {
      avisarError(e);
    }
    return null;
  }

  abrirHoja('Emparejar un dispositivo', [
    parrafo('Emparejá solo dispositivos tuyos: quedan con ', ['control total'], ' de la PC.'),
    seccion('Pedidos pendientes'),
    lista,
    seccion('PIN que muestra Moonlight'),
    pin,
    nombre,
    opcion('Emparejar', 'mas', enviar),
    opcion('Cancelar', 'cerrar', cerrarHoja),
  ]);
  cargar();
  hojaTimer = setInterval(cargar, 2500);
}

async function hojaRegistros() {
  const pre = crear('pre', 'registro', 'Cargando…');
  const cargar = async () => {
    try {
      const { lineas } = await api('registros');
      pre.textContent = '';
      for (const linea of lineas) {
        const clase = /Error|Fatal/.test(linea) ? 'e' : /Warning/.test(linea) ? 'w' : null;
        pre.append(crear('span', clase, `${linea.replace(/^\[\d{4}-\d\d-\d\d /, '[')}\n`));
      }
      pre.scrollIntoView({ block: 'end' });
    } catch (e) {
      pre.textContent = mensajeError(e);
    }
  };
  abrirHoja('Registros de Sunshine', [
    opcion('Actualizar', 'refrescar', cargar),
    pre,
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
  cargar();
}

function hojaTema() {
  const botones = TEMAS.map(t => opcion(t.nombre, 'paleta', () => {
    S.prefs.tema = t.id;
    guardarPrefs();
    aplicarTema();
    botones.forEach((b, i) => b.classList.toggle('activa', TEMAS[i].id === t.id));
    render();
  }, { activa: S.prefs.tema === t.id }));
  abrirHoja('Color del fondo', [...botones, opcion('Listo', 'cerrar', cerrarHoja)]);
}

function hojaMusica() {
  const botones = Object.keys(NIVELES_MUSICA).map(nivel => opcion(NOMBRES_MUSICA[nivel], 'musica', () => {
    S.prefs.musica = nivel;
    guardarPrefs();
    Musica.actualizar();
    botones.forEach((b, i) => b.classList.toggle('activa', Object.keys(NIVELES_MUSICA)[i] === nivel));
    render();
  }, { activa: S.prefs.musica === nivel }));
  abrirHoja('Música de fondo', [
    parrafo('Música ambiental propia que se arma en vivo: ', ['nunca suena igual'], '. Se calla sola cuando empieza un juego.'),
    ...botones,
    opcion('Listo', 'cerrar', cerrarHoja),
  ]);
}

// La app vieja (3.0 a 3.2) le pasa la descarga al navegador, que no tiene la red de la app:
// se la manda por la red de casa, que para los APK está abierta.
function urlDescarga() {
  return EN_APP && !SE_ACTUALIZA_SOLA && S.estado?.ip_casa
    ? `http://${S.estado.ip_casa}:47777/descargas/${APK}` : `descargas/${APK}`;
}

function hojaInstalar() {
  abrirHoja(EN_APP ? 'App CloudGaming' : 'App para Android', [
    EN_APP
      ? parrafo('Tenés la versión ', [window.CloudGaming.version()], `. La más nueva en la PC es la ${VERSION_APP}.`)
      : parrafo('Con la app jugás ', ['desde este menú, sin Moonlight'], ', y se maneja con el joystick.'),
    EN_APP ? null : parrafo('Android te va a pedir permiso para instalar apps de ', ['origen desconocido'], ': dáselo a Chrome solo para esto y después lo podés desactivar.'),
    EN_APP && !SE_ACTUALIZA_SOLA
      ? parrafo('Esta versión baja la actualización con el navegador, que ', ['no llega a la PC'], '. Una sola vez, conectado al wifi de casa, tocá «Descargar de nuevo» o abrí en Chrome ',
        [`http://${S.estado?.ip_casa || 'IP-de-la-PC'}:47777/descargas/${APK}`], '. Desde la 3.3 se actualiza sola.')
      : null,
    SE_ACTUALIZA_SOLA
      ? opcion('Actualizar', 'mas', () => window.CloudGaming.actualizar(`descargas/${APK}`),
        { sub: `${APK} · la baja la app y Android pide confirmar` })
      : opcion(EN_APP ? 'Descargar de nuevo' : 'Descargar la app', 'mas', () => { location.href = urlDescarga(); },
        { sub: `${APK} · se baja con el navegador` }),
    EN_APP ? null : opcion('Versión para la tele', 'mas', () => { location.href = 'descargas/cloudgaming-tv.apk'; },
      { sub: 'cloudgaming-tv.apk · Android TV (32 bits)' }),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

// La app de Android (2.5+) avisa cuando pasa a segundo plano y cuando vuelve, por si el
// WebView no marca la página como oculta: la música no tiene que sonar detrás del juego.
window.panelEnPausa = pausa => {
  Musica.enPausa = pausa;
  if (!pausa) Musica.enJuego = false;
  Musica.actualizar();
};

// La app de Android llama a esto con el botón Atrás: cierra la hoja abierta si la hay.
window.panelAtras = () => {
  if (hoja.hidden) return false;
  cerrarHoja();
  return true;
};

function hojaAcerca() {
  abrirHoja('Panel CloudGaming', [
    parrafo('Panel hecho a medida para manejar Sunshine y WoW Forever desde el celu.'),
    parrafo('Corre en tu PC (puerto ', ['47777'], ') y solo responde a la propia PC y a los equipos de tu red de Tailscale.'),
    parrafo('La página oficial de Sunshine sigue estando en la PC, en https://localhost:47990.'),
    opcion('Volver', 'cerrar', cerrarHoja),
  ]);
}

// ---------------------------------------------------------------- arranque

let timerEstado = 0;

// Sin WoW en la PC (el host no encontró Battle.net, juego: null) su categoría no aparece.
function quitarCategoria(id) {
  const i = CATEGORIAS.findIndex(c => c.id === id);
  if (i < 0) return;
  const actual = CATEGORIAS[S.cat].id;
  CATEGORIAS.splice(i, 1);
  S.cat = Math.max(CATEGORIAS.findIndex(c => c.id === actual), 0);
  construirCategorias();
}

async function actualizar() {
  if (RED_PROPIA && sinRed()) {
    // La red de la app todavía conecta (por ejemplo, al volver a la app): un pedido ahora
    // saldría por la red común y quedaría colgado. El sondeo prueba de nuevo en 3 segundos.
    S.errorEstado = 'sin_conexion';
    pintarIndicadores();
    return;
  }
  try {
    S.estado = await api('estado');
    S.errorEstado = null;
    if (S.estado.juego === null) quitarCategoria('juego');
  } catch (e) {
    if (e.codigo === 'sesion') return;
    S.errorEstado = e.codigo;
  }
  pintarIndicadores();
  if (!$('#menu').hidden) render();
}

function arrancarSondeo() {
  clearInterval(timerEstado);
  actualizar();
  timerEstado = setInterval(() => {
    if (document.visibilityState === 'visible') actualizar();
  }, 3000);
}

document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible' && !$('#menu').hidden) actualizar();
});

function mostrarEntrada(mensaje) {
  clearInterval(timerEstado);
  cerrarHoja();
  $('#menu').hidden = true;
  $('#entrada').hidden = false;
  $('#indicadores').textContent = '';
  $('#error-entrada').textContent = mensaje || '';
}

function mostrarMenu() {
  // La dirección recuerda la categoría: …/#juego abre directo en Juego.
  const inicial = CATEGORIAS.findIndex(c => c.id === location.hash.slice(1));
  if (inicial >= 0) S.cat = inicial;
  $('#entrada').hidden = true;
  $('#menu').hidden = false;
  render(true);
  arrancarSondeo();
  if (new URLSearchParams(location.search).has('instalar')) {
    history.replaceState(null, '', location.pathname + location.hash);
    hojaInstalar();
  }
}

$('#form-entrada').addEventListener('submit', async ev => {
  ev.preventDefault();
  const form = ev.currentTarget;
  const boton = form.querySelector('button');
  boton.disabled = true;
  $('#error-entrada').textContent = '';
  try {
    await api('login', { usuario: form.usuario.value.trim(), clave: form.clave.value });
    form.clave.value = '';
    mostrarMenu();
  } catch (e) {
    $('#error-entrada').textContent = mensajeError(e);
  } finally {
    boton.disabled = false;
  }
});

async function iniciar() {
  aplicarTema();
  setInterval(() => { if (S.prefs.tema === 'hora') aplicarTema(); }, 60000);
  Ondas.iniciar();
  Musica.iniciar();
  construirCategorias();
  pintarReloj();
  setInterval(pintarReloj, 10000);
  $('.avatar').innerHTML = icono('usuario');
  if (S.prefs.ayudaVista) $('#ayuda').classList.add('vista');
  try {
    const { logueado, por_cuenta: porCuenta } = await api('sesion');
    S.porCuenta = Boolean(porCuenta);
    if (logueado) mostrarMenu();
    else mostrarEntrada();
  } catch (e) {
    // Sin conexión (por ejemplo, la red todavía conectando): el menú avisa y se recupera solo cuando
    // vuelva; si para entonces la sesión no sirve, el sondeo lleva a la entrada.
    if (e.codigo === 'sin_conexion') mostrarMenu();
    else mostrarEntrada(mensajeError(e));
  }
}

// Solo en HTTPS (Tailscale serve): permite instalarla como app y abrirla sin conexión.
if ('serviceWorker' in navigator && isSecureContext) {
  navigator.serviceWorker.register('sw.js').catch(() => { /* sin service worker igual anda */ });
}

iniciar();
