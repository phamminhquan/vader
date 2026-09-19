# 1. Add Go's misc/wasm directory to your PATH so Go can find the node runner
#export PATH="$PATH:$(go env GOROOT)/lib/wasm"

# 2. Run your code again
env -i \
  HOME=$HOME \
  PATH="$PATH:$(go env GOROOT)/lib/wasm" \
  GOOS=js \
  GOARCH=wasm \
  go run ./src/ -local

