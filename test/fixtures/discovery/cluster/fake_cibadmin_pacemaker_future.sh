#!/usr/bin/env bash

# Future Pacemaker: stonith-enabled dropped; only fencing-enabled remains in the CIB.
# Checks must use fencing-enabled to find fencing configuration.

cat <<'EOF'
<cib crm_feature_set="4.0.0" validate-with="pacemaker-4.0" epoch="6881" num_updates="0" admin_epoch="0" cib-last-written="Mon Nov 18 17:48:21 2019" update-origin="node01" update-client="crm_attribute" update-user="root" have-quorum="1" dc-uuid="1084783375">
  <configuration>
    <crm_config>
      <cluster_property_set id="cib-bootstrap-options">
        <nvpair id="cib-bootstrap-options-have-watchdog" name="have-watchdog" value="true"/>
        <nvpair id="cib-bootstrap-options-dc-version" name="dc-version" value="4.0.0"/>
        <nvpair id="cib-bootstrap-options-cluster-infrastructure" name="cluster-infrastructure" value="corosync"/>
        <nvpair id="cib-bootstrap-options-cluster-name" name="cluster-name" value="hana_cluster"/>
        <nvpair name="fencing-enabled" value="true" id="cib-bootstrap-options-fencing-enabled"/>
        <nvpair name="placement-strategy" value="balanced" id="cib-bootstrap-options-placement-strategy"/>
      </cluster_property_set>
    </crm_config>
    <nodes/>
    <resources/>
    <constraints/>
  </configuration>
  <status/>
</cib>
EOF
