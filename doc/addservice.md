SELinux: Defining a new init service
------------------------------------

To prevent processes from running in the powerful init SELinux domain,
Android requires that all init spawned processes run in their own dedicated
SELinux sandbox. 


Creating a new service for all Android devices
------------------------------------

To create a new SELinux sandbox for an init spawned service, please do the
following:

1. Create a .te file for your service, similar to:

    system/sepolicy/private/myservice.te
    
    type myservice, domain;
    type myservice_exec, exec_type, file_type;
    typeattribute myservice coredomain;
    
    init_daemon_domain(myservice)
    
2. Ensure that your executable is assigned the correct SELinux label

    system/sepolicy/private/file_contexts
    
    /system/bin/myservice    u:object_r:myservice_exec:s0

3. Rebuild and reflash your device.

Creating a new service for a specific Android device
------------------------------------

Follow the instructions above, but substitute system/sepolicy with
device/MANUFACTURER/HARDWARE/sepolicy and remove the line which contains
"coredomain".


I still get errors after following the steps above.
------------------------------------

1. Verify that your executable is labeled correctly

    $ adb shell
    device:/ $ su
    device:/ # ls -laZ /system/bin/myservice
    -rwxr-xr-x 1 root shell u:object_r:myservice_exec:s0 16312 2017-08-25 17:42 /system/bin/myservice

2. Verify that the policy on the device contains the appropriate transition rules.

    $ adb pull /sys/fs/selinux/policy
    /sys/fs/selinux/policy: 1 file pulled. 10.3 MB/s (358252 bytes in 0.033s)
    $ sesearch --allow -s init -t myservice -c process -p transition ./policy 
    allow init myservice:process { siginh transition rlimitinh };
    $ sesearch --allow -s init -t myservice_exec -c file -p execute ./policy 
    allow init myservice_exec:file { read map getattr open execute };

