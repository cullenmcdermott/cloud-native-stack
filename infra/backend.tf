terraform {
  backend "gcs" {
    bucket = "cns-tf-state"
    prefix = "eidos/prod" # or eidos/dev, eidos/stage, etc.
  }
}