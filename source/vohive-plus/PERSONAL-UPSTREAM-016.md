# 0.1.16 personal: subscription confirmation isolation

Based on the full 0.1.15 personal classic tree. Selectively adapted from
https://github.com/yibaiba/hideck:

- `1598da47796321792b6698c5bb30b42686f36e2b`, BFNOC
  `<15773134+BFNOC@users.noreply.github.com>`: preserve MWI stop state when an
  overlapping Service receives a confirming NOTIFY.
- `9f2cdefb292eca3200082e4208868f59c15f27dc`, yibaiba
  `<946100224@qq.com>`: snapshot shared failure revision at SUBSCRIBE dispatch;
  an old or queued subscription cannot clear a newer Service's failure.

The personal version retains its combined registration/retry store and all
existing identity/event scoping. Its mutex protects a store-wide monotonic
retry revision, including after a confirmed record is deleted. Confirmation
preserves network waits and MWI stop reasons. The diagnostic return value
retains the number of failures cleared before resetting the stored count.

Regression fixtures cover old active/pending NOTIFYs for reg/MWI, queued
refreshes, three MWI stop reasons, and failure-record deletion/recreation.
They use local simulated SIP Services and do not operate physical modems.

No carrier preset, user MWI setting, manual reset API, RF behavior, SA retirement
architecture, UI, dependency, persistent schema or runtime-store lifetime is
changed. Existing third-party copyright and license notices remain applicable.
