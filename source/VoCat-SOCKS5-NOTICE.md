# VoCat SOCKS5 adaptation notice

Copyright (c) 2026 Vocat Project Authors.
Upstream author: MengMengCode <227010654+MengMengCode@users.noreply.github.com>.
Upstream commit: https://github.com/MengMengCode/VoCat/commit/d1968d1c3fb88258d86a6e2e69f9c3304d160711

The SOCKS5 relay DNS error sanitization helper and safe-classification regression in source/vohive-plus/third_party/vowifi-go/engine/ipsec/ are adapted from that commit. Material changes: integrate the helper into the existing personal readSocks5Reply and DecodeSocks5UDPDatagram failure paths; add offline parser regression coverage. No VoCat IKE lifecycle or modem architecture is imported.

The adapted VoCat material retains its upstream Vocat Research & Evaluation License, Version 1.0, reproduced verbatim in third-party-licenses/VoCat-Research-Evaluation-License.md and in the Release license notice asset. This notice does not relicense that material or grant additional rights. Existing project and third-party licenses remain as documented in source/vohive-plus/THIRD_PARTY_NOTICES.md; no license compatibility determination is made here.
