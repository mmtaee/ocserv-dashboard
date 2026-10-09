## copy to vm by rsync
```
rsync -aHAX --delete --info=progress2 \
  --exclude='.git/' \
  --exclude='.env' \
  --exclude='.release' \
  --exclude='node_modules/' \
  --exclude='dist/' \
  --exclude='.yarn/' \
  --exclude='web/admin/node_modules/' \
  --exclude='web/customer/node_modules/' \
  --exclude='web/admin/dist/' \
  --exclude='web/customer/dist/' \
  --exclude='.volume/' \
  --exclude='volumes/' \
  --exclude='logs/' \
  ./ masoud@192.168.0.113:~/ocserv-dashboard/
  
```