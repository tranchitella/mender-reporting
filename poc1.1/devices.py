import uuid
import random
from datetime import datetime

def gen(num, iname, schema):
    for i in range(num):
        if schema == "flat":
            yield random_dev_flat(iname)

        if schema == "flat_typed":
            yield random_dev_flat_typed(iname)

def random_dev_flat(iname):
    uid = uuid.uuid4()
    mac = random_mac()

    d = {
        # needed for bulk insert
        "_index": iname,

        "id": uid.hex,
        "name": "device-" + uid.hex,
        "tenantID": "tenant1",
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow(),
    }

    if random.randint(0, 10) > 5:
        d["status"] = "pending"
    else:
        d["status"] = "accepted"

    d["groupName"] = random.randint(0, 100)

    # identity
    d[attr("identity", "mac")] = mac
    d[attr("identity", "serial_no")] = random.randint(0, 999999999999)

    # inventory
    d[attr("inventory", "mac")] = mac
    d[attr("inventory", "artifact_name")] = "system-M1"
    d[attr("inventory", "device_type")] = "dm1"
    d[attr("inventory", "hostname")] = "Ambarella"
    d[attr("inventory", "ipv4_bcm0")] = "192.168.42.1/24"
    d[attr("inventory", "ipv4_usb0")] = "10.0.1.2/8"
    d[attr("inventory", "ipv4_wlan0")] = "192.168.1.111/24"
    d[attr("inventory", "kernel")] = "Linux version 4.14.181 (charles-chang@rdsuper) (gcc version 8.2.1 20180802 (Linaro GCC 8.2-2018.08~dev)) #1 SMP PREEMPT Fri Mar 12 13:21:16 CST 2021"
    d[attr("inventory", "mac_bcm0")] = mac
    d[attr("inventory", "mac_usb0")] = mac
    d[attr("inventory", "mac_wlan0")] = mac
    d[attr("inventory", "mem_total_kB")] = random.randint(100000, 1000000)
    d[attr("inventory", "mender_bootloader_integration")] = "unknown"
    d[attr("inventory", "mender_client_version")] = "7cb96ca"
    d[attr("inventory", "network_interfaces")] = ["bcm0", "usb0", "wlan0"]
    d[attr("inventory", "os")] = "Ambarella Flexible Linux CV25 (2.5.7) DMS (0.0.0.21B)"
    d[attr("inventory", "rootfs_type")] = "ext4"
    d[attr("inventory", "rootfs_image.checksum")] = "dbc44ce5bd57f0c909dfb15a1efd9fd5d4e426c0fa95f18ea2876e1b8a08818f"
    d[attr("inventory", "rootfs_image.version")] = "system-M1"

    # custom
    d[attr("custom", "tag")] = "value-" + str(random.randint(0,100))

    return d


def random_dev_flat_typed(iname):
    uid = uuid.uuid4()
    mac = random_mac()

    d = {
        # needed for bulk insert
        "_index": iname,

        "id": uid.hex,
        "name": "device-" + uid.hex,
        "tenantID": "tenant1",
        "created_at": datetime.utcnow(),
        "updated_at": datetime.utcnow(),
    }

    if random.randint(0, 10) > 5:
        d["status"] = "pending"
    else:
        d["status"] = "accepted"

    d["groupName"] = random.randint(0, 100)

    # identity
    d[attrs("identity", "mac")] = mac
    d[attrn("identity", "serial_no")] = random.randint(0, 999999999999)

    # inventory
    d[attrs("inventory", "mac")] = mac
    d[attrs("inventory", "artifact_name")] = "system-M1"
    d[attrs("inventory", "device_type")] = "dm1"
    d[attrs("inventory", "hostname")] = "Ambarella"
    d[attrs("inventory", "ipv4_bcm0")] = "192.168.42.1/24"
    d[attrs("inventory", "ipv4_usb0")] = "10.0.1.2/8"
    d[attrs("inventory", "ipv4_wlan0")] = "192.168.1.111/24"
    d[attrs("inventory", "kernel")] = "Linux version 4.14.181 (charles-chang@rdsuper) (gcc version 8.2.1 20180802 (Linaro GCC 8.2-2018.08~dev)) #1 SMP PREEMPT Fri Mar 12 13:21:16 CST 2021"
    d[attrs("inventory", "mac_bcm0")] = mac
    d[attrs("inventory", "mac_usb0")] = mac
    d[attrs("inventory", "mac_wlan0")] = mac
    d[attrn("inventory", "mem_total_kB")] = random.randint(100000, 1000000)
    d[attrs("inventory", "mender_bootloader_integration")] = "unknown"
    d[attrs("inventory", "mender_client_version")] = "7cb96ca"
    d[attrs("inventory", "network_interfaces")] = ["bcm0", "usb0", "wlan0"]
    d[attrs("inventory", "os")] = "Ambarella Flexible Linux CV25 (2.5.7) DMS (0.0.0.21B)"
    d[attrs("inventory", "rootfs_type")] = "ext4"
    d[attrs("inventory", "rootfs_image.checksum")] = "dbc44ce5bd57f0c909dfb15a1efd9fd5d4e426c0fa95f18ea2876e1b8a08818f"
    d[attrs("inventory", "rootfs_image.version")] = "system-M1"

    # custom
    d[attrs("custom", "tag")] = "value-" + str(random.randint(0,100))

    return d

def random_mac():
    mac = [
        random.randint(0x00, 0xFF),
        random.randint(0x00, 0xFF),
        random.randint(0x00, 0xFF),
        random.randint(0x00, 0xFF),
        random.randint(0x00, 0xFF),
        random.randint(0x00, 0xFF),
    ]

    return ":".join(map(lambda x: "%02x" % x, mac))

def attr(scope, name):
    return "{}_{}".format(scope, name)

def attrs(scope, name):
    return "{}_{}_str".format(scope, name)

def attrn(scope, name):
    return "{}_{}_num".format(scope, name)
