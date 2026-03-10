import { apiClient } from "./client";

// Embed configuration types
export interface EmbedConfig {
  enabled: boolean;
  allowed_origins: string[];
  require_api_key: boolean;
  ui_enabled: boolean;
  ui_theme: "light" | "dark" | "auto";
  rate_limit_per_hour: number;
}

export interface EmbedSnippet {
  snippet: string;
  pinned_snippet: string;
}

export interface EmbedAnalytics {
  period: string;
  total_executions: number;
  unique_origins: number;
  origin_stats: Array<{
    origin: string;
    count: number;
  }>;
}

export interface EmbedSnippetParams {
  namespace?: string;
  autoload?: boolean;
  ui?: boolean;
  theme?: "light" | "dark" | "auto";
}

class EmbedApi {
  /**
   * Get embed configuration for a function
   * GET /v1/registry/functions/{author}/{name}/embed
   */
  async getEmbedConfig(author: string, name: string): Promise<EmbedConfig> {
    const response = await apiClient.get<EmbedConfig>(
      `/v1/registry/functions/${author}/${name}/embed`
    );
    return response;
  }

  /**
   * Update embed configuration for a function
   * PUT /v1/registry/functions/{author}/{name}/embed
   */
  async updateEmbedConfig(
    author: string,
    name: string,
    config: Partial<EmbedConfig>
  ): Promise<EmbedConfig> {
    const response = await apiClient.put<EmbedConfig>(
      `/v1/registry/functions/${author}/${name}/embed`,
      config
    );
    return response;
  }

  /**
   * Get embed code snippet for a function
   * GET /v1/registry/functions/{author}/{name}/embed/snippet
   */
  async getEmbedSnippet(
    author: string,
    name: string,
    params?: EmbedSnippetParams
  ): Promise<EmbedSnippet> {
    const queryParams = new URLSearchParams();
    if (params?.namespace) queryParams.append("namespace", params.namespace);
    if (params?.autoload !== undefined)
      queryParams.append("autoload", params.autoload.toString());
    if (params?.ui !== undefined) queryParams.append("ui", params.ui.toString());
    if (params?.theme) queryParams.append("theme", params.theme);

    const url = `/v1/registry/functions/${author}/${name}/embed/snippet${
      queryParams.toString() ? `?${queryParams.toString()}` : ""
    }`;
    return apiClient.get<EmbedSnippet>(url);
  }

  /**
   * Get embed analytics for a function
   * GET /v1/registry/functions/{author}/{name}/embed/analytics
   */
  async getEmbedAnalytics(
    author: string,
    name: string,
    period?: string
  ): Promise<EmbedAnalytics> {
    const queryParams = new URLSearchParams();
    if (period) queryParams.append("period", period);

    const url = `/v1/registry/functions/${author}/${name}/embed/analytics${
      queryParams.toString() ? `?${queryParams.toString()}` : ""
    }`;
    return apiClient.get<EmbedAnalytics>(url);
  }
}

export const embedApi = new EmbedApi();
