require 'json'
require 'net/http'
require 'uri'

module FunctionFly
  # Execution context providing access to host functions.
  class Context
    attr_reader :function_name, :version

    def initialize(api_key: nil, base_url: nil)
      @function_name = ENV["FUNCTIONFLY_FUNCTION_NAME"] || ""
      @version = ENV["FUNCTIONFLY_VERSION"] || ""
      @api_key = api_key || ENV["FUNCTIONFLY_API_KEY"] || ""
      @base_url = base_url || ENV["FUNCTIONFLY_API_URL"] || "https://api.functionfly.com"
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

    # Retrieve an attestation by ID.
    #
    # @param attestation_id [String] The attestation ID (e.g. "att_a1b2c3...")
    # @return [Hash, nil] Attestation data or nil if not found
    def get_attestation(attestation_id)
      response = _api_request(:get, "/v1/trust/attestations/#{attestation_id}")
      response
    rescue StandardError => e
      log("get_attestation failed: #{e.message}")
      nil
    end

    # Delegate execution to another function with trust-aware routing.
    #
    # @param function_id [String] The target function ID
    # @param input [Hash] Input to pass to the target function
    # @param options [Hash] Optional delegation options
    # @option options [Integer] :min_trust_score Minimum trust score (0-100)
    # @option options [String] :min_trust_tier Minimum trust tier
    # @option options [Integer] :timeout_ms Timeout in milliseconds
    # @option options [Boolean] :retry Whether to retry on failure
    # @option options [Integer] :max_retries Maximum retries
    # @return [Hash, nil] Execution result or nil on error
    def delegate(function_id, input, options = {})
      body = { function_id: function_id, input: input }
      body[:options] = options unless options.empty?
      _api_request(:post, "/v1/functions/#{function_id}/execute", body)
    rescue StandardError => e
      log("delegate failed: #{e.message}")
      nil
    end

    private

    def _api_request(method, path, body = nil)
      uri = URI.parse("#{@base_url}#{path}")
      http = Net::HTTP.new(uri.host, uri.port)
      http.use_ssl = uri.scheme == "https"
      http.open_timeout = 10
      http.read_timeout = 30

      case method
      when :get
        req = Net::HTTP::Get.new(uri.request_uri)
      when :post
        req = Net::HTTP::Post.new(uri.request_uri)
        req.body = body.to_json if body
        req["Content-Type"] = "application/json"
      end

      req["Authorization"] = "Bearer #{@api_key}" unless @api_key.empty?

      response = http.request(req)
      return nil unless response.is_a?(Net::HTTPSuccess)

      JSON.parse(response.body)
    rescue StandardError => e
      log("API request failed: #{e.message}")
      nil
    end
  end
end
