resource "kubernetes_namespace" "nuitee" {
  metadata {
    name = "nuitee-challenge"
    labels = {
      app = "nuitee-challenge"
      environment = var.environment
    }
  }
  
 
}

resource "kubernetes_config_map" "api_config" {
  metadata {
    name      = "api-config"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
  }

  data = {
    DATABASE_NAME     = var.database_name
    DATABASE_USER     = var.database_user
    DATABASE_HOST     = "postgres-service"
    DATABASE_PORT     = "5432"
    DATABASE_SSLMODE  = "disable"
    PORT              = "4000"
    ENVIRONMENT       = var.environment
    LITE_API_URL      = var.lite_api_url
  }
}


resource "kubernetes_secret" "api_secrets" {
  metadata {
    name      = "api-secrets"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
  }

  data = {
    DATABASE_PASSWORD = var.database_password
    DATABASE_DSN      = "postgres://${var.database_user}:${var.database_password}@postgres-service:5432/${var.database_name}?sslmode=disable"
    LITE_API_KEY      = var.lite_api_key
  }

  type = "Opaque"
}


resource "kubernetes_persistent_volume" "postgres_pv" {
  metadata {
    name = "postgres-pv"
  }

  spec {
    capacity = {
      storage = var.postgres_storage_size
    }

    access_modes = ["ReadWriteOnce"]
    
    persistent_volume_source {
      host_path {
        path = "/tmp/postgres-data"
        type = "DirectoryOrCreate"
      }
    }

    storage_class_name = "manual"
    persistent_volume_reclaim_policy = "Retain"
  }
  

}


resource "kubernetes_persistent_volume_claim" "postgres_pvc" {
  metadata {
    name      = "postgres-pvc"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.postgres_storage_size
      }
    }
    storage_class_name = "manual"
    volume_name = kubernetes_persistent_volume.postgres_pv.metadata[0].name
  }
}

resource "kubernetes_deployment" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
    labels = {
      app = "postgres"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = {
        app = "postgres"
      }
    }

    template {
      metadata {
        labels = {
          app = "postgres"
        }
      }

      spec {
        container {
          name  = "postgres"
          image = "postgres:16.2-alpine3.18"

          env {
            name  = "POSTGRES_USER"
            value = var.database_user
          }

          env {
            name = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.api_secrets.metadata[0].name
                key  = "DATABASE_PASSWORD"
              }
            }
          }

          env {
            name  = "POSTGRES_DB"
            value = var.database_name
          }

          port {
            container_port = 5432
            name          = "postgres"
          }

          volume_mount {
            name       = "postgres-storage"
            mount_path = "/var/lib/postgresql/data"
          }

          volume_mount {
            name       = "postgres-init"
            mount_path = "/docker-entrypoint-initdb.d"
            read_only  = true
          }

          liveness_probe {
            exec {
              command = ["pg_isready", "-U", var.database_user, "-d", var.database_name]
            }
            initial_delay_seconds = 30
            period_seconds        = 10
          }

          readiness_probe {
            exec {
              command = ["pg_isready", "-U", var.database_user, "-d", var.database_name]
            }
            initial_delay_seconds = 5
            period_seconds        = 5
          }

          resources {
            requests = {
              memory = "256Mi"
              cpu    = "250m"
            }
            limits = {
              memory = "512Mi"
              cpu    = "500m"
            }
          }
        }

        volume {
          name = "postgres-storage"
          persistent_volume_claim {
            claim_name = kubernetes_persistent_volume_claim.postgres_pvc.metadata[0].name
          }
        }

        volume {
          name = "postgres-init"
          config_map {
            name = kubernetes_config_map.postgres_init.metadata[0].name
          }
        }
      }
    }
  }
}


resource "kubernetes_config_map" "postgres_init" {
  metadata {
    name      = "postgres-init"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
  }

  data = {
    "001_init.up.sql" = file("${path.module}/../internal/db/migrations/001_init.up.sql")
  }
}


resource "kubernetes_service" "postgres" {
  metadata {
    name      = "postgres-service"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
    labels = {
      app = "postgres"
    }
  }

  spec {
    selector = {
      app = "postgres"
    }

    port {
      port        = 5432
      target_port = 5432
      name        = "postgres"
    }

    type = "ClusterIP"
  }
}


resource "kubernetes_deployment" "api" {
  metadata {
    name      = "api"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
    labels = {
      app = "api"
    }
  }

  spec {
    replicas = var.api_replicas

    selector {
      match_labels = {
        app = "api"
      }
    }

    template {
      metadata {
        labels = {
          app = "api"
        }
      }

      spec {
        container {
          name  = "api"
          image = "docker.io/madfelps/challenge-nuitee:latest"

          port {
            container_port = 4000
            name          = "http"
          }

          env_from {
            config_map_ref {
              name = kubernetes_config_map.api_config.metadata[0].name
            }
          }

          env_from {
            secret_ref {
              name = kubernetes_secret.api_secrets.metadata[0].name
            }
          }

          liveness_probe {
            http_get {
              path = "/v1/healthcheck"
              port = 4000
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/v1/healthcheck"
              port = 4000
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 3
            failure_threshold     = 3
          }

          resources {
            requests = {
              memory = "128Mi"
              cpu    = "100m"
            }
            limits = {
              memory = "256Mi"
              cpu    = "200m"
            }
          }
        }

        restart_policy = "Always"
      }
    }
  }

  depends_on = [kubernetes_deployment.postgres]
}

resource "kubernetes_service" "api" {
  metadata {
    name      = "api-service"
    namespace = kubernetes_namespace.nuitee.metadata[0].name
    labels = {
      app = "api"
    }
  }

  spec {
    selector = {
      app = "api"
    }

    port {
      port        = 80
      target_port = 4000
      name        = "http"
    }

    type = "ClusterIP"
  }
}
