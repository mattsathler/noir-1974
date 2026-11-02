let audioContext;
let keyNoiseBuffer;
let soundEnabled = true;
const SOUND_STORAGE_KEY = "noir:system:sound-effects:v1";

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

function getAudioContext() {
  if (typeof window === "undefined") return null;
  const AudioContextConstructor = window.AudioContext || window.webkitAudioContext;
  if (!AudioContextConstructor) return null;
  audioContext ??= new AudioContextConstructor();
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
  gain.connect(context.destination);
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
  clack.connect(context.destination);
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
  bodyGain.connect(context.destination);
  bodyNoise.start(bodyStart, Math.random() * 0.015, 0.026);
}

export function playRetroSound(kind = "click", enabled = soundEnabled) {
  if (!enabled) return;
  try {
    const context = getAudioContext();
    if (!context) return;
    const compose = () => {
      const startAt = context.currentTime + 0.01;
      if (kind === "boot") {
        [[261.6, 0], [329.6, 0.12], [392, 0.24], [523.3, 0.39]].forEach(([frequency, offset]) => {
          playTone(context, frequency, startAt + offset, 0.24, 0.15, frequency * 0.985, "triangle");
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
