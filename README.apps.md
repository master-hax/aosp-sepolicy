The policy defines multiple types and attributes for apps. This document is a
high-level overview of these. For further details on each types, refer to their
specific files in the public/ and private/ directories.

## appdomain
In general, all apps will have the `appdomain` attribute. You can think of
`appdomain` as any app started by Zygote.

## untrusted_app
Third-party apps (for example, installed from the Play Store), will have an
`untrusted_app_xx` type where xx is the targetSdkVersion. For instance, an app
with targetSdkVersion = 32 will be typed as `untrusted_app_32`. Not all
targetSdkVersion have a specific type, some version are skipped when no
difference where introduced (see public/untrusted_app.te for more details).
Apps targetting the current sdk, will be typed as `untrusted_app`.

The `untrusted_app_all` attribute can be used to reference all the types
described in this section (that is, `untrusted_app_*`).

## isolated_app
Apps may be restricted when using isolatedProcess=true in their manifest. In
this case, they will be assigned the `isolated_app` type. A similar type
`isolated_compute_app` exist for some specific services.

Both types `isolated_app` and `isolated_compute_app` are grouped under the
attribute `isolated_app_all`.

## ephemeral_app & sdk_sandbox

## unrestricted_app
This is an attribute to reference any app that is not isolated, ephemeral nor
sdk_sandbox.

## system_app
Apps pre-installed on a device and running with the system UID.

## priv_app
Apps shipped as part of the device, by the device manufacturer.


