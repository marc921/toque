resource "scaleway_k8s_cluster" "cluster" {
  name    = "toque-cluster"
  type    = "kapsule"
  version = "1.30.2"
  cni     = "cilium"
  private_network_id = scaleway_vpc_private_network.pn.id
  delete_additional_resources = true
}

resource "scaleway_k8s_pool" "pool" {
  cluster_id = scaleway_k8s_cluster.cluster.id
  name       = "toque-pool"
  node_type  = "PLAY2-NANO"
  size       = 1
  autohealing = true
}

resource "local_file" "kubeconfig" {
    content  = scaleway_k8s_cluster.cluster.kubeconfig.0.config_file
    filename = "kubeconfig.yaml"
}