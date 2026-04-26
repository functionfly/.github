module FunctionFly
  # Base class for FunctionFly functions.
  # Subclass this and implement the #handle method.
  #
  #   class MyFunction < FunctionFly::Function
  #     def handle(input, ctx)
  #       '{"message": "Hello from Ruby!"}'
  #     end
  #   end
  #
  #   FunctionFly.run(MyFunction)
  class Function
    # Process the input and return the output as a string.
    # @param input [String] The function input (JSON string)
    # @param ctx [FunctionFly::Context] Execution context
    # @return [String] The function output (JSON string)
    def handle(input, ctx)
      raise NotImplementedError, "Subclass must implement #handle"
    end
  end
end
