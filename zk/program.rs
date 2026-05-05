// SP1 ZK Program - executes inside zkVM
#![no_main]
sp1_zkvm::entrypoint!(main);

use sha2::{Sha256, Digest};

pub fn main() {
    // Read batch data from prover
    let batch_id = sp1_zkvm::io::read::<String>();
    let payments_count = sp1_zkvm::io::read::<u32>();
    let total_amount = sp1_zkvm::io::read::<u64>();
    let state_root = sp1_zkvm::io::read::<String>();
    
    // Compute hash (in production: verify all signatures + state transitions)
    let input = format!("{}-{}-{}-{}", batch_id, payments_count, total_amount, state_root);
    let hash = Sha256::digest(input.as_bytes());
    
    // Output the result
    sp1_zkvm::io::commit(&hash[..16]);
}
