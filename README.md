# prometheus-http-sd-sakuracloud

さくらのクラウドのサーバー・ロードバランサアプライアンスの情報をもとにターゲット情報を作成し、prometheus の http-sd によるサービスディスカバリーに応答します。

## tag

サーバー・ロードバランサにタグを設定することでサービスディスカバリーの対象の判定を行います。

| タグ                        | 説明                                                                                                                                     |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `prometheus.io/scrape=true` | サービスディスカバリーの対象とする                                                                                                       |
| `prometheus.io/port=<port>` | ターゲットのポートを`<port>`に変更する (デフォルトは 9100)                                                                               |
| `sd/exclude=<vip>:<port>`   | ロードバランサにおいて除外する vip:port を指定します。複数の vip:port を指定する場合は、`sd/exclude=<ip>:<port>`のタグを複数指定します。 |

## prometheus.yml 例

```yaml

---
scrape_configs:
  - job_name: "sacloud_servers"
    http_sd_configs:
      - url: "http://localhost:8080/discovery/server"

  - job_name: "sacloud_loadbalancers"
    http_sd_configs:
      - url: "http://localhost:8080/discovery/loadbalancer"
    relabel_configs:
      - source_labels:
          - __meta_loadbalancer_name
        target_label: lb_name
      - source_labels:
          - __meta_loadbalancer_vip_port
        target_label: lb_vip_port
```
