package tee_attestation

import rego.v1

default allow := false
default hw_verified := false
default image_digest_verified := false
default nonce_verified := false
default issuer_verified := false
default secboot_verified := false
default sw_name_verified := false

allow if {
	hw_verified
	image_digest_verified
	nonce_verified
	issuer_verified
	secboot_verified
	sw_name_verified
}

hw_verified if input.hwmodel in ["GCP-TDX", "TDX"]
image_digest_verified if input.submods.container.image_digest in data.allowed_image_digests
issuer_verified if input.iss == "https://confidentialcomputing.googleapis.com"
secboot_verified if input.secboot == true
sw_name_verified if input.swname == "CONFIDENTIAL_SPACE"
nonce_verified if {
	input.eat_nonce == data.expected_nonce
}