#include <core.p4>
#include <v1model.p4>

typedef bit<48> EthernetAddress;

header ethernet_t {
    EthernetAddress dstAddr;
    EthernetAddress srcAddr;
    bit<16>         etherType;
}

header payment_h {
    bit<32> batch_id;
    bit<32> payment_id;
    bit<32> from_account;
    bit<32> to_account;
    bit<64> amount;
    bit<8>  is_valid;
}

struct headers {
    ethernet_t  ethernet;
    payment_h   payment;
}

struct metadata {
    bit<64> total_amount;
    bit<32> tx_count;
}

parser MyParser(packet_in packet, out headers hdr, inout metadata meta, inout standard_metadata_t std_meta) {
    state start {
        packet.extract(hdr.ethernet);
        packet.extract(hdr.payment);
        transition accept;
    }
}

control MyIngress(inout headers hdr, inout metadata meta, inout standard_metadata_t std_meta) {
    action do_aggregate() {
        meta.total_amount = meta.total_amount + hdr.payment.amount;
        meta.tx_count = meta.tx_count + 1;
    }
    
    action do_pass() {
        // No changes
    }
    
    table aggregate_table {
        key = {
            hdr.payment.is_valid: exact;
        }
        actions = {
            do_aggregate;
            do_pass;
        }
        size = 2;
        default_action = do_pass();
    }
    
    apply {
        aggregate_table.apply();
    }
}

control MyEgress(inout headers hdr, inout metadata meta, inout standard_metadata_t std_meta) {
    apply {
        // Update amount field with total for response
        hdr.payment.amount = meta.total_amount;
    }
}

control MyDeparser(packet_out packet, in headers hdr) {
    apply {
        packet.emit(hdr.ethernet);
        packet.emit(hdr.payment);
    }
}

control MyVerifyChecksum(inout headers hdr, inout metadata meta) {
    apply {}
}

control MyComputeChecksum(inout headers hdr, inout metadata meta) {
    apply {}
}

V1Switch(
    MyParser(),
    MyVerifyChecksum(),
    MyIngress(),
    MyEgress(),
    MyComputeChecksum(),
    MyDeparser()
) main;
