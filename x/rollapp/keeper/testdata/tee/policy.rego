package confidential_space

import rego.v1

default allow := false
default hw_verified := false
default image_digest_verified := false
default audience_verified := false
default nonce_verified := false
default issuer_verified := false
default secboot_verified := false
default sw_name_verified := false

allow if {
	hw_verified
	image_digest_verified
	audience_verified
	nonce_verified
	issuer_verified
	secboot_verified
	sw_name_verified
}

hw_verified if input.hwmodel in data.allowed_hwmodel
image_digest_verified if input.submods.container.image_digest in data.allowed_submods_container_image_digest
audience_verified if input.aud in data.allowed_aud
issuer_verified if input.iss in data.allowed_issuer
secboot_verified if input.secboot in data.allowed_secboot
sw_name_verified if input.swname in data.allowed_sw_name
nonce_verified if {
	input.eat_nonce == "%s"
}