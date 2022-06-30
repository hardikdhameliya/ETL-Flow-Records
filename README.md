(This project does not use Go Modules)

## ETL pipeline

This project was created to read flow records, enrich the records based on oui, and upload the enriched data into a database.

All operations have customizable parameters, which are listed in the `config.yaml` file.

The next section provides detailed information on the operations.

### Fetch

This action retrieves flow data from the file on a regular basis. It deletes the file from the disk once the complete file has been read.

The sample flow data is shown below.

```json
{
  "id":"38657394411904521",
  "app_id":1183,
  "app_name":"amazon_aws",
  "service":"ssl",
  "ip_version":4,
  "src_ip":"10.155.10.108",
  "src_port":25365,
  "dst_ip":"30.30.23.68",
  "dst_port":443,
  "protocol":6,
  "start_time":"2020-03-18T22:10:47.498Z",
  "end_time":"2020-03-18T22:10:47.702Z",
  "end_reason":2,
  "tcp_flags":2,
  "src_mac":"12:22:0a:9b:0a:6c",
  "dst_mac":"12:22:1e:1e:17:44",
  "src_bytes":1394,
  "src_packets":15,
  "dst_bytes":31720,
  "dst_packets":23
}
```

### Enrich

This operation enriches flow data by using OUI (Organizationally Unique Identifier). On the [IEEE website](https://standards-oui.ieee.org/oui/oui.txt), you can find examples of OUI information. The vendor's OUI is cached in Redis. OUI data is obtained from Redis and appended to the flow record for each flow event based on `src_mac` and `dst_mac`.

### Upload

This operation writes enriched flow entries into the database. MongoDB is used to store the final data records.


