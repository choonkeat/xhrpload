# xhrpload

A minimal HTTP server for transferring files from a browser to another computer over the network. Useful when you only have access to a web browser on one end, but still want a fast, no‑dependency upload tool.

---

## What This Does

`xhrpload` lets you start a simple web server locally that can receive file uploads via XHR (XMLHttpRequest) or `<form>` POST. It’s intended for situations where you cannot use command‑line tools like `nc`, `scp`, or `rsync` — for example, when the sending device is restricted to a browser‑only environment.

**Example usage:**

```bash
go run -output ~/Downloads
```

Then open `http://127.0.0.1:8080` in a browser to upload files (Change `127.0.0.1` to the reciever ip address). The uploaded files will be written to the configured output directory.

---

## Why This Repo Exists

Sometimes, both sender and receiver are not equally capable:

* The **sender** only has a browser (e.g., mobile, kiosk, VM, or sandboxed host).
* The **receiver** can run command‑line tools.

`xhrpload` fills that gap by exposing a simple HTTP endpoint for file uploads, viewable and usable from any browser without extra setup.

When **both** sides have shell access and can open ports, you can skip HTTP entirely and use command‑line alternatives for faster direct transfers — see below.

---

## Alternative: Using `nc` (Netcat)

If both machines can use the command line and reach each other via TCP, `nc` (netcat) transfers files *much faster* than browser uploads.

Let’s call the two machines:

* **A** → sender (has the file)
* **B** → receiver (will get the file)
* Port: `9000` (change if needed)

---

### 1. Single File

**Receiver (B):**

```bash
nc -l 9000 > received.bin
# or: nc -l -p 9000 > received.bin
```

**Sender (A):**

```bash
nc B.hostname.or.ip 9000 < bigfile.bin
```

Show progress:

```bash
pv bigfile.bin | nc B 9000
```

Verify integrity:

```bash
sha256sum bigfile.bin
sha256sum received.bin
```

---

### 2. Whole Directory

**Receiver (B):**

```bash
nc -l 9000 | tar -C /destination/path -xvf -
```

**Sender (A):**

```bash
tar -C /path/to/parent -cf - mydir | nc B 9000
```

With compression:

```bash
tar -C /path/to/parent -cf - mydir | gzip -1 | nc B 9000
# receiver:
nc -l 9000 | gzip -d | tar -C /destination/path -xvf -
```

Show progress:

```bash
tar -C /path/to/parent -cf - mydir | pv | nc B 9000
```

---

### 3. Reverse Direction (if only outbound allowed)

If **B** cannot listen, have **A** listen instead:

```bash
# On A (sender)
nc -l 9000 < bigfile.bin
# On B (receiver)
nc A.hostname.or.ip 9000 > received.bin
```

---

### 4. Via SSH Tunnel (when inbound ports are blocked)

If B can SSH to A:

```bash
# On B
ssh -N -R 9000:localhost:9000 user@A
# Then on B
nc -l 9000 > received.bin
# On A
nc localhost 9000 < bigfile.bin
```

---

### 5. Using `ncat` (with TLS)

```bash
# Receiver
ncat --ssl -l 9000 > received.bin
# Sender
ncat --ssl B 9000 < bigfile.bin
```

Or explicitly send/recv:

```bash
ncat --send-only B 9000 < bigfile.bin
ncat --recv-only -l 9000 > received.bin
```

---

## Notes

* If `nc -l 9000` errors, try `nc -l -p 9000`.
* Some versions hang after EOF — add `-q 1`.
* Use `sha256sum` to verify integrity.
* Tar preserves permissions and timestamps.
* Netcat often saturates a gigabit link easily.

---

### Example Full Command

```bash
# Receiver (B)
nc -l 9000 | tar -xvf - -C ~/Downloads

# Sender (A)
tar -cf - myproject | pv | nc B 9000
```

---

## Summary

* Use **`xhrpload`** if the sender only has a browser.
* Use **`nc`/`tar`** if both machines have a shell and network reachability.

Both aim for simplicity and speed — one through HTTP, the other through direct TCP.
