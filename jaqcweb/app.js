
const yearEl = document.getElementById("year");
const btnUpdateUI = document.getElementById("btnUpdateUI");
const btnEdit = document.getElementById("btnEdit");
const btnSave = document.getElementById("btnSave");
const btnCancel = document.getElementById("btnCancel");
const statusEl = document.getElementById("status");

// Display-only fields
const serial    = document.getElementById("serial");
const hwClass   = document.getElementById("hwClass");
const hwVersion = document.getElementById("hwVersion");
const fwVersion = document.getElementById("fwVersion");

// Editable fields
const wifiSsid  = document.getElementById("wifiSsid");
const wifiPass  = document.getElementById("wifiPass");
const mqttUri   = document.getElementById("mqttUri");
const mqttUser  = document.getElementById("mqttUser");
const mqttPass  = document.getElementById("mqttPass");

let editMode = false;

function setStatus(msg, kind = "info") {
    statusEl.textContent = msg;
    statusEl.style.borderColor = (kind === "error") ? "var(--danger)" :
        (kind === "ok") ? "var(--accent)" : "var(--border)";
}

async function updateUIFiles() {
    try {
        const res = await fetch("/api/update_web_all");
        if (!res.ok) throw new Error("update UI failed");
        location.reload(true);
    } catch (e) {
        setStatus("could not fetch UI files", "error");
    }
}

function setEditMode(on) {
    editMode = !!on;
    // Toggle buttons
    btnUpdateUI.hidden= !on;
    btnEdit.hidden  =  on;
    btnSave.hidden  = !on;
    btnCancel.hidden= !on;

    // Enable/disable editable inputs
    document.querySelectorAll("[data-editable]").forEach(inp => {
        inp.disabled = !on;
    });

    setStatus(on ? "Edit mode enabled. Make changes and click Save." : "Edit mode exited.");
}

function loadLocal() {
    // Fill display-only from placeholders (you can wire to /api/status later)
    serial.value    = localStorage.getItem("serial") || "";
    hwClass.value   = localStorage.getItem("hwClass") || "";
    hwVersion.value = localStorage.getItem("hwVersion") || "";
    // fwVersion.value = localStorage.getItem("fwVersion") || "";

    // Load editable from localStorage for test convenience
    wifiSsid.value = localStorage.getItem("wifiSsid") || "";
    wifiPass.value = localStorage.getItem("wifiPass") || "";
    mqttUri.value  = localStorage.getItem("mqttUri")  || "";
    mqttUser.value = localStorage.getItem("mqttUser") || "";
    mqttPass.value = localStorage.getItem("mqttPass") || "";
}

async function loadFromEspStatus() {
// Optional: wire to your ESP endpoint
// try {
//   const res = await fetch("/api/status");
//   if (!res.ok) throw new Error("status fetch failed");
//   const j = await res.json();
//   // we only get ssid/ip from your current impl; fill what we can
//   wifiSsid.value = j.ssid || wifiSsid.value;
//   setStatus(`Connected: ${j.state || ""} ${j.ip ? " (" + j.ip + ")" : ""}`, "ok");
// } catch (e) {
//   setStatus("Could not fetch status from device.", "error");
// }
}

async function saveChanges() {
    // Simple local save for now; wire publish to cmd/config later
    localStorage.setItem("serial", serial.value);
    localStorage.setItem("hwClass", hwClass.value);
    localStorage.setItem("hwVersion", hwVersion.value);
    // localStorage.setItem("fwVersion", fwVersion.value);

    localStorage.setItem("wifiSsid", wifiSsid.value);
    localStorage.setItem("wifiPass", wifiPass.value);

    localStorage.setItem("mqttUri",  mqttUri.value);
    localStorage.setItem("mqttUser", mqttUser.value);
    localStorage.setItem("mqttPass", mqttPass.value);

    // Example: send to ESP for STA connect (uncomment when ready)
    // try {
    //   const r1 = await fetch("/api/connect", {
    //     method: "POST",
    //     headers: { "Content-Type": "application/json" },
    //     body: JSON.stringify({ ssid: wifiSsid.value, pass: wifiPass.value })
    //   });
    //   if (!r1.ok) throw new Error("Wi‑Fi connect failed");

    //   // If you add an MQTT config endpoint later:
    //   // const r2 = await fetch("/api/mqtt_config", {
    //   //   method: "POST",
    //   //   headers: { "Content-Type": "application/json" },
    //   //   body: JSON.stringify({ uri: mqttUri.value, user: mqttUser.value, pass: mqttPass.value })
    //   // });
    //   // if (!r2.ok) throw new Error("MQTT config failed");

    //   setStatus("Settings saved. Device updating…", "ok");
    // } catch (e) {
    //   setStatus("Save failed: " + e.message, "error");
    //   return;
    // }

    setEditMode(false);
    setStatus("Settings saved locally.", "ok");
}

function cancelChanges() {
    setEditMode(false);
    loadLocal();  // revert to stored values
    setStatus("Changes canceled.");
}

btnUpdateUI.addEventListener("click", () => {
    console.log('updateing UI files...');
    updateUIFiles();
});
btnEdit.addEventListener("click", () => {
    console.log("editing...");
    setEditMode(true);
});
btnSave.addEventListener("click", saveChanges);
btnCancel.addEventListener("click", cancelChanges);

// init
setEditMode(false);
loadLocal();
// loadFromEspStatus(); // enable when the endpoint is ready
