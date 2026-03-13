#!/usr/bin/env node

// run the script in /build folder based on current architecture
// if the current architecture is not supported, throw an error
// and exit the process

const { spawn } = require("child_process");
const { join } = require('path');
const os = require("os");

const platform = os.platform();
const arch = os.arch();

const supportedEnvs = [
  'linux/x64',
  'darwin/x64',
  'darwin/arm64',
  'linux/arm64',
  'win32/x64',
]

const supported = supportedEnvs.includes(`${platform}/${arch}`);

if (!supported) {
  throw new Error(`Unsupported platform: ${platform} ${arch}`);
}

const mapNodejsArchToGo = {
  x64: "amd64",
  arm64: "arm64",
};

const mapPlatformToGo = {
  linux: "linux",
  darwin: "darwin",
  win32: "windows",
};

const script = `main.go-${mapPlatformToGo[platform]}-${mapNodejsArchToGo[arch]}${platform === "win32" ? ".exe" : ""}`;

const binPath = join(__dirname, "..", "build", script);
const allArgs = process.argv.slice(2).join(" ");

// execute the script with all the arguments in interactive mode
// so that the user can interact with the script
const child = spawn(`"${binPath}" ${allArgs}`, { stdio: "inherit", shell: true });

child.on("error", (err) => {
  console.error(err);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    // If the process was terminated by a signal, use exit code 1
    process.exit(1);
  }
  // If code is null/undefined, treat as failure
  process.exit(typeof code === "number" ? code : 1);
});


