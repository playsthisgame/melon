#!/usr/bin/env node
'use strict'

const { execFileSync } = require('child_process')
const path = require('path')

const binaryName = process.platform === 'win32' ? 'mln.exe' : 'mln'
const binary = path.join(__dirname, '..', binaryName)

try {
  execFileSync(binary, process.argv.slice(2), { stdio: 'inherit' })
} catch (e) {
  process.exit(e.status || 1)
}
