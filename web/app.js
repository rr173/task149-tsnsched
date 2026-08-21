document.querySelectorAll("button[data-api]").forEach((button) => {
  button.addEventListener("click", async () => {
    const output = document.getElementById("out");
    try {
      const response = await fetch(button.dataset.api);
      output.textContent = JSON.stringify(await response.json(), null, 2);
    } catch (error) {
      output.textContent = String(error);
    }
  });
});
