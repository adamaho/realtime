let msgCount = 0;

const clientCount = 16;
async function client(clientId: number) {
  const response = await fetch("https://localhost:3000/count");

  if (!response.ok) {
    return;
  }

  const reader = response.body?.getReader({ mode: undefined });
  while (true) {
    if (!reader) {
      console.log("no reader?");
      break;
    }

    // @ts-ignore
    const { done, value } = await reader.read();

    if (done) {
      console.log("done?");
      break;
    }

    const decoder = new TextDecoder("utf-8");
    const str = decoder.decode(value.buffer);
    msgCount += 1;

    const totalMsgs = msgCount / clientCount;
    console.log(totalMsgs);
  }

  return;
}

async function run() {
  const clients: Promise<void>[] = [];
  for (let i = 0; i < clientCount; i++) {
    clients.push(client(i + 1));
  }

  await Promise.all(clients);
}

run();
