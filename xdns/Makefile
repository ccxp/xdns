#
# Copyright (C) 2016 Canonical.com
#
# This is free software, licensed under the GNU General Public License v2.
# See /LICENSE for more information.
#

include $(TOPDIR)/rules.mk

PKG_NAME:=xdns
PKG_VERSION:=1.1.5
PKG_RELEASE:=1

PKG_SOURCE:=xdns-$(PKG_VERSION).tar.xz
PKG_MAINTAINER:=xpchen <xpchen@qq.com>
PKG_LICENSE:=GPL

include $(INCLUDE_DIR)/package.mk

define Package/xdns
  SUBMENU:=Utilities
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=http/socks/dns proxy
endef

define Package/xdns/description
 http proxy
endef

define Build/Prepare
	$(CP) ./src/* $(PKG_BUILD_DIR)/
endef

define Build/Configure
endef

GOARCH:=$(ARCH)

ifeq ($(GOARCH),i386)
  GOARCH:=386
  ifeq ($(CONFIG_CPU_TYPE),"pentium4")
    GOSUBARCH:=GO386=sse2
  else
    GOSUBARCH:=GO386=387
  endif
endif
ifeq ($(GOARCH),x86_64)
  GOARCH:=amd64
endif
ifeq ($(GOARCH),aarch64)
  GOARCH:=arm64
endif
ifeq ($(GOARCH),arm)
  ifeq ($(CONFIG_arm_v5),y)
    GOSUBARCH:=GOARM=5
  endif
  ifeq ($(CONFIG_arm_v6),y)
    GOSUBARCH:=GOARM=6
  endif
  ifeq ($(CONFIG_arm_v7),y)
    GOSUBARCH:=GOARM=7
  endif
endif

#GOENV:=GOPATH=$(PKG_BUILD_DIR) GOARCH=$(GOARCH) $(GOSUBARCH) CGO_ENABLED=0 CC=$(TARGET_CC)
GOENV:=GOARCH=$(GOARCH) $(GOSUBARCH) CGO_ENABLED=0 CC=$(TARGET_CC)
GOFLAGS:=--ldflags='-s -w'

define Build/Compile
	$(GOENV) go build -o $(PKG_BUILD_DIR)/xdns $(GOFLAGS) $(PKG_BUILD_DIR)/*.go
endef

define Package/xdns/install
	$(INSTALL_DIR) $(1)/usr/bin $(1)/etc/init.d
	$(INSTALL_BIN) $(PKG_BUILD_DIR)/xdns $(1)/usr/bin/
	$(INSTALL_BIN) ./files/xdns.init $(1)/etc/init.d/xdns
endef

$(eval $(call BuildPackage,xdns))
