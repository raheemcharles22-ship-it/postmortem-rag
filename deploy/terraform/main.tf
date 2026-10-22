resource "kubernetes_secret" "app_secrets" {
  metadata {
    name = "app-secrets"
  }
  data = {
    "openai-key"  = var.openai_api_key
    "db-password" = var.db_password
  }
}

resource "kubernetes_deployment" "app" {
  metadata {
    name = "app"
  }
  spec {
    replicas = 1
    selector {
      match_labels = { app = "rag" }
    }
    template {
      metadata {
        labels = { app = "rag" }
      }
      spec {
        container {
          name              = "app"
          image             = var.app_image
          image_pull_policy = "IfNotPresent"
          port {
            container_port = 8080
          }
          env {
            name = "DB_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.app_secrets.metadata[0].name
                key  = "db-password"
              }
            }
          }
          env {
            name  = "DATABASE_URL"
            value = "postgres://postgres:$(DB_PASSWORD)@postgres:5432/postmortem_rag?sslmode=disable"
          }
          env {
            name = "OPENAI_API_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.app_secrets.metadata[0].name
                key  = "openai-key"
              }
            }
          }
          readiness_probe {
            http_get {
              path = "/"
              port = 8080
            }
            initial_delay_seconds = 5
            period_seconds        = 5
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "app" {
  metadata {
    name = "app"
  }
  spec {
    selector = { app = "rag" }
    port {
      port        = 8080
      target_port = 8080
    }
  }
}

resource "kubernetes_persistent_volume_claim" "postgres_data" {
  wait_until_bound = false 
  metadata {
    name = "postgres-data"
  }
  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = "1Gi"
      }
    }
  }
}

resource "kubernetes_deployment" "postgres" {
  metadata {
    name = "postgres"
  }
  spec {
    replicas = 1
    selector {
      match_labels = { app = "postgres" }
    }
    template {
      metadata {
        labels = { app = "postgres" }
      }
      spec {
        volume {
          name = "postgres-data"
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim.postgres_data.metadata[0].name
          }
        }

        container {
          name              = "postgres"
          image             = "pgvector/pgvector:pg17"
          image_pull_policy = "IfNotPresent"

          port {
            container_port = 5432
          }

          env {
            name  = "POSTGRES_DB"
            value = "postmortem_rag"
          }
          env {
            name  = "POSTGRES_USER"
            value = "postgres"
          }
          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.app_secrets.metadata[0].name
                key  = "db-password"
              }
            }
          }

          volume_mount {
            name       = "postgres-data"
            mount_path = "/var/lib/postgresql/data"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "postgres" {
  metadata {
    name = "postgres"
  }
  spec {
    selector = { app = "postgres" }
    port {
      port        = 5432
      target_port = 5432
    }
  }
}