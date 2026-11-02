let audioContext;
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

function playTone(context, frequency, startAt, duration, volume, endFrequency = frequency) {
  const oscillator = context.createOscillator();
  const gain = context.createGain();
  oscillator.type = "square";
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

export function playRetroSound(kind = "click") {
  if (!soundEnabled) return;
  try {
    const context = getAudioContext();
    if (!context) return;
    const compose = () => {
      const startAt = context.currentTime + 0.01;
      if (kind === "notification") {
        [[659, 0], [880, 0.075], [1175, 0.15]].forEach(([frequency, offset]) => {
          playTone(context, frequency, startAt + offset, 0.07, 0.38, frequency * 0.94);
        });
        return;
      }
      if (kind === "key") {
        const pitch = 560 + Math.floor(Math.random() * 5) * 47;
        playTone(context, pitch, startAt, 0.022, 0.22, pitch * 0.72);
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
