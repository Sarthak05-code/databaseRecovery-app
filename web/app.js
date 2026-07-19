async function triggerAction(endpoint) {
  const logBox = document.getElementById("outputLog");

  // Gather all inputs
  const payload = {
    host: document.getElementById("host").value,
    port: document.getElementById("port").value,
    user: document.getElementById("user").value,
    password: document.getElementById("password").value,
    dbname: document.getElementById("dbname").value,
    storageDir: document.getElementById("storageDir").value,
    backupFilePath: document.getElementById("backupFilePath").value,
    slackWebhook: document.getElementById("slackWebhook").value,
  };

  if (!payload.dbname) {
    logBox.innerText =
      "Validation Error: Database Name parameter is mandatory.";
    return;
  }

  if (endpoint === "/api/restore" && !payload.backupFilePath) {
    logBox.innerText =
      "Validation Error: Target Restore File Path is required for restoration operations.";
    return;
  }

  logBox.innerText = `[${new Date().toLocaleTimeString()}] Contacting backend binary execution runtime... Please wait.`;

  try {
    const response = await fetch(endpoint, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    const textResult = await response.text();

    if (response.ok) {
      const data = JSON.parse(textResult);
      logBox.innerText = `[SUCCESS] ${data.message}\n${data.elapsed ? `Execution Duration: ${data.elapsed}` : ""}`;
    } else {
      logBox.innerText = `[ERROR TYPE ${response.status}] ${textResult}`;
    }
  } catch (err) {
    logBox.innerText = `Network/Runtime connectivity failure: ${err.message}`;
  }
}
