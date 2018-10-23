load test_helper

@test "zones (when none exist)" {
  run $ew zones

  fixture="$(cat tests/fixtures/zones_none)"

  [ "${output}" = "${fixture}" ]
}

@test "zone-create (with connection)" {
  run $ew zone-create \
    --name "ok." \
    --email "test@test.com" \
    --admin-group-id "ok-group" \
    --zone-connection-key-name "vinyldns." \
    --zone-connection-key "nzisn+4G2ldMn0q1CV3vsg==" \
    --zone-connection-primary-server "vinyldns-bind9" \
    --transfer-connection-key-name "vinyldns." \
    --transfer-connection-key "nzisn+4G2ldMn0q1CV3vsg==" \
    --transfer-connection-primary-server "vinyldns-bind9"

  fixture="$(cat tests/fixtures/zone_create_connection)"

  [ "${output}" = "${fixture}" ]
}

@test "zone-create (with no connection)" {
  run $ew zone-create \
    --name "vinyldns." \
    --email "admin@test.com" \
    --admin-group-id "ok-group"

  fixture="$(cat tests/fixtures/zone_create_no_connection)"

  [ "${output}" = "${fixture}" ]
}

@test "zone-create (with invalid zone connection)" {
  run $ew zone-create \
    --name "ok-invalid-connection." \
    --email "test@test.com" \
    --admin-group-id "ok-group" \
    --zone-connection-key "nzisn+4G2ldMn0q1CV3vsg==" \
    --zone-connection-primary-server "vinyldns-bind9"

  fixture="$(cat tests/fixtures/zone_create_invalid_zone_connection)"

  [ "${status}" -eq 1 ]
  [ "${output}" = "${fixture}" ]
}

@test "zone-create (with invalid transfer connection)" {
  run $ew zone-create \
    --name "ok-invalid-connection." \
    --email "test@test.com" \
    --admin-group-id "ok-group" \
    --transfer-connection-key "nzisn+4G2ldMn0q1CV3vsg==" \
    --transfer-connection-primary-server "vinyldns-bind9"

  fixture="$(cat tests/fixtures/zone_create_invalid_transfer_connection)"

  [ "${status}" -eq 1 ]
  [ "${output}" = "${fixture}" ]
}

@test "groups" {
  run $ew groups

  fixture="$(cat tests/fixtures/groups)"

  [ "${output}" = "${fixture}" ]
}

@test "group" {
  run $ew group --group-id ok-group

  fixture="$(cat tests/fixtures/group)"

  [ "${output}" = "${fixture}" ]
}

@test "group-members" {
  # `| sed '10,$d'` to get only the first 10 lines of output,
  # as the last line contains a `Created` field that will differ
  # with each startup of Dockerized VinylDNS.
  run $ew group-members --group-id ok-group | sed '10,$d'

  fixture="$(cat tests/fixtures/group_members)"

  [ "${output}" = "${fixture}" ]
}