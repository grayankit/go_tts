let isPaused = false;
const elevenVoices = [
  { id: "pNInz6obpgDQGcFmaJgB", name: "Boris (ElevenLabs)" },
  { id: "EXAVITQu4vr4xnSDxMaL", name: "Bella" },
  { id: "ErXwobaYiN019PkySvjV", name: "Antoni" },
  { id: "21m00Tcm4TlvDq8ikWAM", name: "Rachel" },
  { id: "AZnzlk1XvdvUeBnXmlld", name: "Domi" },
];

async function loadVoices() {
  const select = document.getElementById("voice");
  select.innerHTML = "";

  // ElevenLabs group
  const optgroup1 = document.createElement("optgroup");
  optgroup1.label = "ElevenLabs Voices";
  elevenVoices.forEach(voice => {
    const option = document.createElement("option");
    option.value = `eleven:${voice.id}`;
    option.textContent = voice.name;
    optgroup1.appendChild(option);
  });
  select.appendChild(optgroup1);

  // espeak-ng group
  const res = await fetch("/api/voices");
  const text = await res.text();
  const lines = text.trim().split("\n").slice(1);
  const optgroup2 = document.createElement("optgroup");
  optgroup2.label = "Espeak-NG Voices";

  lines.forEach(line => {
    const parts = line.trim().split(/\s+/);
    const lang = parts[1];
    const name = parts[3];
    const option = document.createElement("option");
    option.value = lang;
    option.textContent = `${lang} - ${name}`;
    optgroup2.appendChild(option);
  });
  select.appendChild(optgroup2);
}

// Fetches the audio data and plays it.
async function fetchAndPlay(text, voice) {
  if (!text) return;
  const res = await fetch("/api/tts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text, voice }),
  });
  const blob = await res.blob();
  const audioUrl = URL.createObjectURL(blob);
  const player = document.getElementById("player");
  player.src = audioUrl;
  return new Promise((resolve) => {
    player.onended = () => resolve();
    player.onerror = () => resolve();
    player.play();
  })
}

// Called by the "Speak" button and history items.
// Sends the text to the backend to be broadcast or queued.
async function sendText() {
  const text = document.getElementById("text").value;
  const voice = document.getElementById("voice").value;
  if (!text) return;

  const res = await fetch("/api/speak", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text, voice }),
  });

  if (res.ok && isPaused) {
    refreshQueue();
  }
}

async function previewVoice() {
  const voice = document.getElementById("voice").value;
  const previewText = "This is a voice preview.";

  // Preview calls TTS directly, bypassing the queue.
  fetchAndPlay(previewText, voice);
}

async function fetchHistory() {
  const res = await fetch("/api/history");
  const history = await res.json();
  const list = document.getElementById("history");
  list.innerHTML = "";
  history.slice().reverse().forEach(item => {
    const li = document.createElement("li");
    li.textContent = item;
    li.onclick = () => {
      document.getElementById("text").value = item;
      sendText(); // This now correctly calls the queueing endpoint.
    };
    list.appendChild(li);
  });
}

const evtSource = new EventSource("/events");

// This is called when the server sends a message via SSE.
evtSource.onmessage = async (e) => {
  const text = e.data;
  const voice = document.getElementById("voice").value;
  document.getElementById("text").value = text;
  await fetchAndPlay(text, voice);
};

async function togglePause() {
  const res = await fetch('/api/pause', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ paused: !isPaused }),
  });
  if (res.ok) {
    const data = await res.json();
    isPaused = data.paused;
    updatePauseButton();
    if (isPaused) {
      refreshQueue();
    } else {
      document.getElementById('queue-list').innerHTML = '';
    }
  }
}

function updatePauseButton() {
  const button = document.querySelector('button[onclick="togglePause()"]');
  if (isPaused) {
    button.textContent = 'Play';
    button.classList.remove('bg-gray-600', 'hover:bg-gray-700');
    button.classList.add('bg-yellow-500', 'hover:bg-yellow-600');
  } else {
    button.textContent = 'Pause';
    button.classList.add('bg-gray-600', 'hover:bg-gray-700');
    button.classList.remove('bg-yellow-500', 'hover:bg-yellow-600');
  }
}

async function refreshQueue() {
  const res = await fetch('/api/queue');
  console.log(res)
  const data = await res.json();
  console.log(data)
  const ul = document.getElementById('queue-list');
  ul.innerHTML = '';

  data.forEach((item) => {
    const li = document.createElement('li');
    if (typeof item === 'string') {
      li.textContent = item;
    } else {
      li.textContent = `${item.voice}: ${item.text}`;
    }
    ul.appendChild(li);
  });
}

async function exportAudio() {
  const text = document.getElementById("text").value;
  const voice = document.getElementById("voice").value;
  if (!text) return;

  const res = await fetch("/api/tts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ text, voice }),
  });

  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = "tts.wav";
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}


document.addEventListener('DOMContentLoaded', () => {
  loadVoices();
  updatePauseButton();
});
setInterval(() => {
  if (isPaused) {
    refreshQueue();
  }
}, 2000);
