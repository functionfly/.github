// FunctionFly example: Hello from Ruby
// Manifest: {"name":"hello-ruby","version":"1.0.0","runtime":"ruby","entry":"hello.rb"}

def handler(input)
  '{"message": "Hello from Ruby!", "input_length": ' + input.length.to_s + ', "ok": true}'
end

# Read input and execute
input = $stdin.read
puts handler(input)
