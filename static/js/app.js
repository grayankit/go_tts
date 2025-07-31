let isPaused = false;
const queue = [];
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

  async function sendText() {
    const text = document.getElementById("text").value;
    const voice = document.getElementById("voice").value;
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
    player.play();
  }

  async function previewVoice() {
    const voice = document.getElementById("voice").value;
    const previewText = "This is a voice preview.";
    const player = document.getElementById("player");

    if (voice.startsWith("eleven:")) {
      const res = await fetch("/api/tts", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text: previewText, voice }),
      });
      const blob = await res.blob();
      player.src = URL.createObjectURL(blob);
    } else {
      player.src = `/api/preview?voice=${voice}`;
    }
    player.play();
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
        sendText();
      };
      list.appendChild(li);
    });
  }
  const evtSource = new EventSource("/events");

evtSource.onmessage = (e) => {
  const text = e.data;
    if(isPaused){
      queue.push(text)
        refreshQueue();
    }else{
      playText(text)
    }
}
function togglePause() {
  isPaused = !isPaused;
  if (!isPaused) {
    while (queue.length > 0) {
      const text = queue.shift();
      playText(text);
    }
  }
}
  function playText(text) {
  const voice = document.getElementById("voice").value;
  document.getElementById("text").value = text;
  sendText(); // your playback function
}
    async function refreshQueue() {
  const res = await fetch('/api/queue');
  const data = await res.json(); // array of text or {text, voice}
      console.log(data)

  const ul = document.getElementById('queue-list');
  ul.innerHTML = '';

  data.forEach((item, idx) => {
    const li = document.createElement('li');
    if (typeof item === 'string') {
      li.textContent = item;
    } else {
      li.textContent = `${item.voice}: ${item.text}`;
    }
    ul.appendChild(li);
  });
}




  loadVoices();