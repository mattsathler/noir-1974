let audioContext;
let masterGain;
let keyNoiseBuffer;
let soundEnabled = true;
const SOUND_STORAGE_KEY = "noir:system:sound-effects:v1";
const SOUND_VOLUME_STORAGE_KEY = "noir:system:sound-volume:v1";
const DEFAULT_SOUND_VOLUME = 1;

export function readRetroSoundVolume() {
  if (typeof window === "undefined") return DEFAULT_SOUND_VOLUME;
  try {
    const saved = window.localStorage.getItem(SOUND_VOLUME_STORAGE_KEY);
    if (saved === null) return DEFAULT_SOUND_VOLUME;
    const value = Number(saved);
    return Number.isFinite(value) ? Math.max(0, Math.min(1, value)) : DEFAULT_SOUND_VOLUME;
  } catch {
    return DEFAULT_SOUND_VOLUME;
  }
}

let soundVolume = readRetroSoundVolume();

export function setRetroSoundVolume(volume) {
  soundVolume = Number.isFinite(Number(volume)) ? Math.max(0, Math.min(1, Number(volume))) : DEFAULT_SOUND_VOLUME;
  if (audioContext && masterGain) masterGain.gain.setTargetAtTime(soundVolume, audioContext.currentTime, 0.015);
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(SOUND_VOLUME_STORAGE_KEY, String(soundVolume));
  } catch {
    // Keep the volume control available for this session if storage is unavailable.
  }
}

export function readRetroSoundEnabled() {
  if (typeof window === "undefined") return true;
  try {
    return window.localStorage.getItem(SOUND_STORAGE_KEY) !== "false";
  } catch {
    return true;
  }
}

export function setRetroSoundEnabled(enabled) {
  soundEnabled = Boolean(enabled);
  if (typeof window === "undefined") return;
  try {
    window.localStorage.setItem(SOUND_STORAGE_KEY, String(soundEnabled));
  } catch {
    // Keep sound controls available for this session if storage is unavailable.
  }
}

export function unlockRetroAudio() {
  if (!readRetroSoundEnabled()) return;
  try {
    const context = getAudioContext();
    if (context?.state === "suspended") context.resume().catch(() => {});
  } catch {
    // Audio unlock must never block sign-in or account creation.
  }
}

function getAudioContext() {
  if (typeof window === "undefined") return null;
  const AudioContextConstructor = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextConstructor) return null;
  audioContext ??= new AudioContextConstructor();
  if (!masterGain) {
    masterGain = audioContext.createGain();
    masterGain.gain.value = soundVolume;
    masterGain.connect(audioContext.destination);
  }
  return audioContext;
}

function playTone(context, frequency, startAt, duration, volume, endFrequency = frequency, waveform = "square") {
  const oscillator = context.createOscillator();
  const gain = context.createGain();
  oscillator.type = waveform;
  oscillator.frequency.setValueAtTime(frequency, startAt);
  oscillator.frequency.exponentialRampToValueAtTime(Math.max(1, endFrequency), startAt + duration);
  gain.gain.setValueAtTime(0.0001, startAt);
  gain.gain.exponentialRampToValueAtTime(volume, startAt + 0.003);
  gain.gain.exponentialRampToValueAtTime(0.0001, startAt + duration);
  oscillator.connect(gain);
  gain.connect(masterGain || context.destination);
  oscillator.start(startAt);
  oscillator.stop(startAt + duration + 0.005);
}

function playTypewriterKey(context, startAt) {
  if (!keyNoiseBuffer) {
    const duration = 0.06;
    keyNoiseBuffer = context.createBuffer(1, Math.ceil(context.sampleRate * duration), context.sampleRate);
    const samples = keyNoiseBuffer.getChannelData(0);
    for (let index = 0; index < samples.length; index++) samples[index] = Math.random() * 2 - 1;
  }

  const noise = context.createBufferSource();
  const highPass = context.createBiquadFilter();
  const lowPass = context.createBiquadFilter();
  const clack = context.createGain();
  noise.buffer = keyNoiseBuffer;
  highPass.type = "highpass";
  highPass.frequency.setValueAtTime(340 + Math.random() * 140, startAt);
  lowPass.type = "lowpass";
  lowPass.frequency.setValueAtTime(2600 + Math.random() * 700, startAt);
  clack.gain.setValueAtTime(0.0001, startAt);
  clack.gain.exponentialRampToValueAtTime(0.3, startAt + 0.0015);
  clack.gain.exponentialRampToValueAtTime(0.0001, startAt + 0.038);
  noise.connect(highPass);
  highPass.connect(lowPass);
  lowPass.connect(clack);
  clack.connect(masterGain || context.destination);
  noise.start(startAt, Math.random() * 0.015, 0.038);

  const bodyNoise = context.createBufferSource();
  const bodyLowPass = context.createBiquadFilter();
  const bodyHighPass = context.createBiquadFilter();
  const bodyGain = context.createGain();
  const bodyStart = startAt + 0.004;
  bodyNoise.buffer = keyNoiseBuffer;
  bodyLowPass.type = "lowpass";
  bodyLowPass.frequency.setValueAtTime(850 + Math.random() * 250, bodyStart);
  bodyHighPass.type = "highpass";
  bodyHighPass.frequency.setValueAtTime(100 + Math.random() * 50, bodyStart);
  bodyGain.gain.setValueAtTime(0.0001, bodyStart);
  bodyGain.gain.exponentialRampToValueAtTime(0.2, bodyStart + 0.001);
  bodyGain.gain.exponentialRampToValueAtTime(0.0001, bodyStart + 0.026);
  bodyNoise.connect(bodyHighPass);
  bodyHighPass.connect(bodyLowPass);
  bodyLowPass.connect(bodyGain);
  bodyGain.connect(masterGain || context.destination);
  bodyNoise.start(bodyStart, Math.random() * 0.015, 0.026);
}

export function playRetroSound(kind = "click", enabled = soundEnabled) {
  if (!enabled || soundVolume <= 0) return;
  try {
    const context = getAudioContext();
    if (!context) return;
    const compose = () => {
      if (soundVolume <= 0) return;
      const startAt = context.currentTime + 0.01;
      if (kind === "boot") {
        [[261.6, 0], [329.6, 0.15], [392, 0.3], [523.3, 0.48]].forEach(([frequency, offset]) => {
          playTone(context, frequency, startAt + offset, 0.34, 0.34, frequency * 0.985, "triangle");
        });
        return;
      }
      if (kind === "warning") {
        [[310, 0], [220, 0.11]].forEach(([frequency, offset]) => {
          playTone(context, frequency, startAt + offset, 0.085, 0.24, frequency * 0.82);
        });
        return;
      }
      if (kind === "notification") {
        [[659, 0], [880, 0.075], [1175, 0.15]].forEach(([frequency, offset]) => {
          playTone(context, frequency, startAt + offset, 0.07, 0.38, frequency * 0.94);
        });
        return;
      }
      if (kind === "key") {
        playTypewriterKey(context, startAt);
        return;
      }
      playTone(context, 420, startAt, 0.032, 0.3, 190);
    };
    if (context.state === "suspended") {
      context.resume().then(compose).catch(() => {});
      return;
    }
    compose();
  } catch {
    // Audio is an enhancement; unavailable or blocked audio must not affect play.
  }
}
