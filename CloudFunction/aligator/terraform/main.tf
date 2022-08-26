terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "3.8.0"
    }
  }
}

provider "google" {
  credentials = file("C:/Users/AbdelmadjidYOUS/Documents/GitHub/AH584312/keys/ah584312-5947cab34099.json")

  project = "ah584312"
  region  = "eu-west1"
  zone    = "eu-west1-c"
}


# Create new storage bucket
resource "google_storage_bucket" "static" {
  name          = "aligator"
  location      = "EUROPE-WEST1"
  storage_class = "REGIONAL"
}


# Create new storage bucket
resource "google_storage_bucket" "static-config" {
  name          = "aligator-param"
  location      = "EUROPE-WEST1"
  storage_class = "REGIONAL"
  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  
  cors {
    origin          = ["*"]
    method          = ["OPTIONS","GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }
}

resource "google_pubsub_topic" "load_bq" {
  name = "load_bq"
  message_storage_policy {
    allowed_persistence_regions = [
      "europe-west1",
    ]
  }

}


resource "google_storage_bucket" "static-site" {
  name                        = "aligator-web"
  location                    = "EUROPE-WEST1"
  storage_class               = "REGIONAL"
# uniform_bucket_level_access = true
  #  force_destroy = true

  #  uniform_bucket_level_access = true

  website {
    main_page_suffix = "index.html"
    not_found_page   = "404.html"
  }
  cors {
    origin          = ["*"]
    method          = ["OPTIONS","GET", "HEAD", "PUT", "POST", "DELETE"]
    response_header = ["*"]
    max_age_seconds = 3600
  }

}