#include <core.p4>
#include <v1model.p4>

header payment_h {
    bit<32> batch_id;
    bit<32> payment_id;
    bit<32> from_account;
    bit<32> to_account;
    bit<64> amount;
    bit<8>  is_valid;
}

struct headers {
    payment_h payment;
}

struct metadata {}

parser MyParser(packet_in packet, out headers hdr, inout metadata meta, inout standard_metadata_t std_meta) {
    state start {
        packet.extract(hdr.payment);
        transition accept;
    }
}

control MyIngress(inout headers hdr, inout metadata meta, inout standard_metadata_t std_meta) {
    Register<bit<64>, bit<32>>(128) batch_sums;
    Register<bit<32>, bit<32>>(128) batch_counts;
    
    action aggregate() {
        bit<64> current_sum;
        bit<32> current_count;
        
        batch_sums.read(current_sum, hdr.payment.batch_id);
        batch_counts.read(current_count, hdr.payment.batch_id);
        
        current_sum = current_sum + hdr.payment.amount;
        current_count = current_count + 1;
        
        batch_sums.write(hdr.payment.batch_id, current_sum);
        batch_counts.write(hdr.payment.batch_id, current_count);
    }
    
    table do_aggregate {
        key = {
            hdr.payment.is_valid: exact;
        }
        actions = {
            aggregate;
            NoAction;
        }
        size = 2;
        default_action = NoAction();
    }
    
    apply {
        do_aggregate.apply();
    }
}

control MyEgress(inout headers hdr, inout metadata meta, inout standard_metadata_t std_meta) {
    apply {}
}

control MyDeparser(packet_out packet, in headers hdr) {
    apply {
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
