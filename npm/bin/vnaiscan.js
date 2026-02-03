#!/usr/bin/env node
/**
 * vnaiscan npx wrapper
 * 
 * Usage: npx @vnai-dev/vnaiscan scan image:tag
 * 
 * This wrapper runs vnaiscan via Docker to ensure all tools are available.
 * For native binary, use the install script instead.
 */

const { spawn } = require('child_process');
const { existsSync } = require('fs');

const DOCKER_IMAGE = 'ghcr.io/vnai-dev/vnaiscan:latest';

function hasDocker() {
  try {
    require('child_process').execSync('docker --version', { stdio: 'ignore' });
    return true;
  } catch {
    return false;
  }
}

function runViaDocker(args) {
  const dockerArgs = [
    'run', '--rm',
    '-v', '/var/run/docker.sock:/var/run/docker.sock',
    '-v', `${process.cwd()}:/workspace`,
    '-w', '/workspace',
    DOCKER_IMAGE,
    ...args
  ];

  const child = spawn('docker', dockerArgs, {
    stdio: 'inherit',
    env: process.env
  });

  child.on('error', (err) => {
    console.error('Failed to run Docker:', err.message);
    console.error('\nMake sure Docker is running and you have access to the Docker socket.');
    process.exit(1);
  });

  child.on('exit', (code) => {
    process.exit(code || 0);
  });
}

function main() {
  const args = process.argv.slice(2);

  if (!hasDocker()) {
    console.error('Error: Docker is required to run vnaiscan via npx.');
    console.error('\nAlternatives:');
    console.error('  1. Install Docker: https://docs.docker.com/get-docker/');
    console.error('  2. Install natively: curl -sSL https://scan.vnai.dev/install.sh | sh');
    process.exit(1);
  }

  // Pull image if needed (first run)
  if (args[0] !== '--skip-pull') {
    console.log('Ensuring vnaiscan image is available...');
    try {
      require('child_process').execSync(`docker pull ${DOCKER_IMAGE}`, { 
        stdio: ['ignore', 'ignore', 'inherit'] 
      });
    } catch {
      // Image might already exist, continue
    }
  }

  runViaDocker(args);
}

main();
