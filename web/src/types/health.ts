export interface HealthResponse {
  status: string
  services: Record<string, string>
}
