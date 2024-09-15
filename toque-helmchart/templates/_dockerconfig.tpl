{
  "auths": {
    "{{ .Values.scaleway.containerRegistry }}": {
      "username": "nologin",
      "password": "{{ .Values.scaleway.secretKey }}",
      "auth": "{{ printf "nologin:%s" .Values.scaleway.secretKey | b64enc }}"
    }
  }
}
