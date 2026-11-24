# 0.1.19-personal-classic

Based on complete 0.1.18 personal source. Adapt HiDeck 7b1c7413a77834c4bedc03d63dedd05920037b2e (ilazycat, 2026-11-20): recognize WAP SI/SL and MMS payloads before the shared 2948 port fallback; explicit recognized content types override conflicting payload hints. Preserve raw data, OMA fallback and SIM OTA handling. Apply the same isolated fix to the byte-identical embedded VoWiFi SMS classifier, with the same full wire regressions. No modem lifecycle, SMS sending/storage, network, carrier or UI changes.

Upstream classifier and regression diff retained; personal tests cover content-type/hint precedence, raw/port preservation and full 8-bit SMS-DELIVER decoding. Baseline tests demonstrate the shared-port bug in both decoders. These are simulated protocol tests; no real SMS, calls or notification-delivery acceptance is claimed.

ML307 discovery/AT inbox/identity, phone media/mode, SIM lifecycle and earlier architecture/carrier changes remain deferred for separate adaptation and authorized hardware acceptance. Original authors and licenses retained.
