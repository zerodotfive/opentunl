<script setup>
</script><!-- Recommended OpenVPN clients per platform (client page) -->
<template>
  <div style="padding: 24px 24px 0">
    <a-typography-title :level="5" style="margin: 0; font-size: 18px">Get a client</a-typography-title>
    <a-typography-text type="secondary">Install an OpenVPN client, then import the config downloaded below.</a-typography-text>
    <a-row :gutter="[16, 16]" style="margin-top: 16px">
      <a-col :xs="24" :md="7">
        <a-card style="height: 100%">
          <template #title><WindowsOutlined /> Windows</template>
          <a-typography-paragraph>
            <a-typography-text strong>OpenVPN Connect</a-typography-text><br />
            The official client from OpenVPN Inc.
          </a-typography-paragraph>
          <a-button type="primary" href="https://apps.microsoft.com/detail/xpdlgw32hhrb5p" target="_blank" rel="noopener">
            <template #icon><DownloadOutlined /></template>Microsoft Store
          </a-button>
        </a-card>
      </a-col>
      <a-col :xs="24" :md="7">
        <a-card style="height: 100%">
          <template #title><AppleOutlined /> macOS</template>
          <a-typography-paragraph>
            <a-typography-text strong>Tunnelblick</a-typography-text><br />
            Free open source client for macOS.
          </a-typography-paragraph>
          <a-button type="primary" href="https://tunnelblick.net/" target="_blank" rel="noopener">
            <template #icon><DownloadOutlined /></template>tunnelblick.net
          </a-button>
        </a-card>
      </a-col>
      <a-col :xs="24" :md="10">
        <a-card style="height: 100%">
          <template #title><CodeOutlined /> Linux</template>
          <a-typography-paragraph style="margin-bottom: 8px">Install the <a-typography-text code>openvpn</a-typography-text> package:</a-typography-paragraph>
          <a-typography-text type="secondary">Debian, Ubuntu</a-typography-text>
          <a-typography-paragraph code copyable style="margin-bottom: 8px">sudo apt install openvpn</a-typography-paragraph>
          <a-typography-text type="secondary">Fedora, RHEL</a-typography-text>
          <a-typography-paragraph code copyable style="margin-bottom: 8px">sudo dnf install openvpn</a-typography-paragraph>
          <a-typography-text type="secondary">Connect</a-typography-text>
          <a-typography-paragraph code copyable style="margin-bottom: 16px">sudo openvpn --config client.ovpn</a-typography-paragraph>
          <a-typography-paragraph style="margin-bottom: 0">Or import config to network settings of your desktop environment</a-typography-paragraph>
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>