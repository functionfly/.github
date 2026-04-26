module FunctionFly
  # Execution context providing access to host functions.
  class Context
    attr_reader :function_name, :version

    def initialize
      @function_name = ENV["FUNCTIONFLY_FUNCTION_NAME"] || ""
      @version = ENV["FUNCTIONFLY_VERSION"] || ""
    end

    # Get an environment variable.
    def get_env(key)
      ENV[key]
    end

    # Log a message.
    def log(msg)
      $stderr.puts "[functionfly] #{msg}"
    end

    # Return an error response.
    def error(code, message)
      '{"error": {"code": "' + code + '", "message": "' + message.gsub('"', '\\"') + '"}}'
    end

    # Return a success JSON response.
    def json(data)
      '{"ok": true, "data": ' + data.to_s + '}'
    end
  end
end
