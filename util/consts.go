// Package util provides reused functions and constants
package util

const ProgramName = "pb"

const DefaultPort = 2850

const EnvVarServer = "PB_CLIPBOARD_SERVER"
const EnvVarPort = "PB_CLIPBOARD_PORT"
const EnvVarKey = "PB_CLIPBOARD_KEY"

const HeaderFingerprint = "X-PB-Key-Fingerprint"
const HeaderSignature = "X-PB-Signature"

const RequestCopy = "/copy"
const RequestPaste = "/paste"
const RequestOpen = "/open"
const RequestQuit = "/quit"
