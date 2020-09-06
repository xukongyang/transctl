package transrpc

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	var torrents []Torrent
	if err := json.NewDecoder(strings.NewReader(torrentJSON)).Decode(&torrents); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(torrents) != 2 {
		t.Errorf("expected 2 torrents, got: %d", len(torrents))
	}
	var session Session
	if err := json.NewDecoder(strings.NewReader(sessionJSON)).Decode(&session); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

const torrentJSON = `[
  {
    "activityDate": 0,
    "addedDate": 1573007237,
    "bandwidthPriority": 0,
    "comment": "a comment",
    "corruptEver": 0,
    "creator": "a creator",
    "dateCreated": 1526135894,
    "desiredAvailable": 0,
    "doneDate": 0,
    "downloadDir": "/media/downloads/",
    "downloadLimit": 100,
    "downloadLimited": false,
    "downloadedEver": 0,
    "error": 0,
    "errorString": "",
    "eta": -1,
    "etaIdle": -1,
    "fileStats": [
      {
        "bytesCompleted": 2433,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 100000000,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 100000000,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 100000000,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 100000000,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 100000000,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 36433243,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 100000000,
        "priority": 0,
        "wanted": true
      },
      {
        "bytesCompleted": 371,
        "priority": 0,
        "wanted": true
      }
    ],
    "files": [
      {
        "bytesCompleted": 2433,
        "length": 2433,
        "name": ""
      },
      {
        "bytesCompleted": 100000000,
        "length": 100000000,
        "name": ""
      },
      {
        "bytesCompleted": 100000000,
        "length": 100000000,
        "name": ""
      },
      {
        "bytesCompleted": 100000000,
        "length": 100000000,
        "name": ""
      },
      {
        "bytesCompleted": 100000000,
        "length": 100000000,
        "name": ""
      },
      {
        "bytesCompleted": 100000000,
        "length": 100000000,
        "name": ""
      },
      {
        "bytesCompleted": 36433243,
        "length": 36433243,
        "name": ""
      },
      {
        "bytesCompleted": 100000000,
        "length": 100000000,
        "name": ""
      },
      {
        "bytesCompleted": 371,
        "length": 371,
        "name": ""
      }
    ],
    "hashString": "ffffffffffffffffffffffffffffffffffffffff",
    "haveUnchecked": 0,
    "haveValid": 636436047,
    "honorsSessionLimits": true,
    "id": 15,
    "isFinished": false,
    "isPrivate": true,
    "isStalled": true,
    "leftUntilDone": 0,
    "magnetLink": "",
    "manualAnnounceTime": -1,
    "maxConnectedPeers": 50,
    "metadataPercentComplete": 1,
    "name": "",
    "peer-limit": 50,
    "peers": [],
    "peersConnected": 0,
    "peersFrom": {
      "fromCache": 0,
      "fromDht": 0,
      "fromIncoming": 0,
      "fromLpd": 0,
      "fromLtep": 0,
      "fromPex": 0,
      "fromTracker": 0
    },
    "peersGettingFromUs": 0,
    "peersSendingToUs": 0,
    "percentDone": 1,
    "pieceCount": 607,
    "pieceSize": 1048576,
    "pieces": "/////////////////////////////////////////////////////////////////////////////////////////////////////g==",
    "priorities": [
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0,
      0
    ],
    "queuePosition": 14,
    "rateDownload": 0,
    "rateUpload": 0,
    "recheckProgress": 0,
    "secondsDownloading": 0,
    "secondsSeeding": 95645,
    "seedIdleLimit": 30,
    "seedIdleMode": 0,
    "seedRatioLimit": 2,
    "seedRatioMode": 0,
    "sizeWhenDone": 636436047,
    "startDate": 1573175949,
    "status": 6,
    "torrentFile": "",
    "totalSize": 636436047,
    "trackerStats": [
      {
        "announce": "https://",
        "announceState": 1,
        "downloadCount": -1,
        "hasAnnounced": true,
        "hasScraped": true,
        "host": "https://",
        "id": 0,
        "isBackup": false,
        "lastAnnouncePeerCount": 0,
        "lastAnnounceResult": "Success",
        "lastAnnounceStartTime": 1573262463,
        "lastAnnounceSucceeded": true,
        "lastAnnounceTime": 1573262464,
        "lastAnnounceTimedOut": false,
        "lastScrapeResult": "Tracker gave HTTP response code 414 (Request-URI Too Long)",
        "lastScrapeStartTime": 0,
        "lastScrapeSucceeded": true,
        "lastScrapeTime": 1573262464,
        "lastScrapeTimedOut": 0,
        "leecherCount": -1,
        "nextAnnounceTime": 1573264264,
        "nextScrapeTime": 1573264270,
        "scrape": "https://",
        "scrapeState": 1,
        "seederCount": 2,
        "tier": 0
      }
    ],
    "trackers": [
      {
        "announce": "https://",
        "id": 0,
        "scrape": "https://",
        "tier": 0
      }
    ],
    "uploadLimit": 100,
    "uploadLimited": false,
    "uploadRatio": 0,
    "uploadedEver": 0,
    "wanted": [
      1,
      1,
      1,
      1,
      1,
      1,
      1,
      1,
      1
    ],
    "webseeds": [],
    "webseedsSendingToUs": 0
  },
  {
    "activityDate": 0,
    "addedDate": 1573007345,
    "bandwidthPriority": 0,
    "comment": "a comment",
    "corruptEver": 0,
    "creator": "a creator",
    "dateCreated": 1516659693,
    "desiredAvailable": 0,
    "doneDate": 0,
    "downloadDir": "/media/downloads/",
    "downloadLimit": 100,
    "downloadLimited": false,
    "downloadedEver": 0,
    "error": 0,
    "errorString": "",
    "eta": -1,
    "etaIdle": -1,
    "fileStats": [
      {
        "bytesCompleted": 2864864807,
        "priority": 0,
        "wanted": true
      }
    ],
    "files": [
      {
        "bytesCompleted": 2864864807,
        "length": 2864864807,
        "name": ""
      }
    ],
    "hashString": "ffffffffffffffffffffffffffffffffffffffff",
    "haveUnchecked": 0,
    "haveValid": 2864864807,
    "honorsSessionLimits": true,
    "id": 600,
    "isFinished": false,
    "isPrivate": true,
    "isStalled": true,
    "leftUntilDone": 0,
    "magnetLink": "",
    "manualAnnounceTime": -1,
    "maxConnectedPeers": 50,
    "metadataPercentComplete": 1,
    "name": "",
    "peer-limit": 50,
    "peers": [],
    "peersConnected": 0,
    "peersFrom": {
      "fromCache": 0,
      "fromDht": 0,
      "fromIncoming": 0,
      "fromLpd": 0,
      "fromLtep": 0,
      "fromPex": 0,
      "fromTracker": 0
    },
    "peersGettingFromUs": 0,
    "peersSendingToUs": 0,
    "percentDone": 1,
    "pieceCount": 684,
    "pieceSize": 4194304,
    "pieces": "//////////////////////////////////////////////////////////////////////////////////////////////////////////////////A=",
    "priorities": [
      0
    ],
    "queuePosition": 595,
    "rateDownload": 0,
    "rateUpload": 0,
    "recheckProgress": 0,
    "secondsDownloading": 0,
    "secondsSeeding": 95109,
    "seedIdleLimit": 30,
    "seedIdleMode": 0,
    "seedRatioLimit": 2,
    "seedRatioMode": 0,
    "sizeWhenDone": 2864864807,
    "startDate": 1573175949,
    "status": 6,
    "torrentFile": "",
    "totalSize": 2864864807,
    "trackerStats": [
      {
        "announce": "https://",
        "announceState": 1,
        "downloadCount": 576,
        "hasAnnounced": true,
        "hasScraped": true,
        "host": "https://",
        "id": 0,
        "isBackup": false,
        "lastAnnouncePeerCount": 0,
        "lastAnnounceResult": "Success",
        "lastAnnounceStartTime": 1573262457,
        "lastAnnounceSucceeded": true,
        "lastAnnounceTime": 1573262459,
        "lastAnnounceTimedOut": false,
        "lastScrapeResult": "Tracker gave HTTP response code 414 (Request-URI Too Long)",
        "lastScrapeStartTime": 1573255251,
        "lastScrapeSucceeded": true,
        "lastScrapeTime": 1573262459,
        "lastScrapeTimedOut": 0,
        "leecherCount": 0,
        "nextAnnounceTime": 1573264259,
        "nextScrapeTime": 1573264260,
        "scrape": "https://",
        "scrapeState": 1,
        "seederCount": 30,
        "tier": 0
      }
    ],
    "trackers": [
      {
        "announce": "https://",
        "id": 0,
        "scrape": "https://",
        "tier": 0
      }
    ],
    "uploadLimit": 100,
    "uploadLimited": false,
    "uploadRatio": 0,
    "uploadedEver": 0,
    "wanted": [
      1
    ],
    "webseeds": [],
    "webseedsSendingToUs": 0
  }
]`

const sessionJSON = `{
  "alt-speed-down": 50,
  "alt-speed-enabled": false,
  "alt-speed-time-begin": 540,
  "alt-speed-time-day": 127,
  "alt-speed-time-enabled": false,
  "alt-speed-time-end": 1020,
  "alt-speed-up": 50,
  "blocklist-enabled": false,
  "blocklist-size": 0,
  "blocklist-url": "http://www.example.com/blocklist",
  "cache-size-mb": 4,
  "config-dir": "/var/lib/transmission-daemon/.config/transmission-daemon",
  "dht-enabled": true,
  "download-dir": "/media/downloads",
  "download-dir-free-space": 31366467584,
  "download-queue-enabled": true,
  "download-queue-size": 4,
  "encryption": "preferred",
  "idle-seeding-limit": 30,
  "idle-seeding-limit-enabled": false,
  "incomplete-dir": "/media/downloads",
  "incomplete-dir-enabled": false,
  "lpd-enabled": false,
  "peer-limit-global": 200,
  "peer-limit-per-torrent": 50,
  "peer-port": 51413,
  "peer-port-random-on-start": false,
  "pex-enabled": true,
  "port-forwarding-enabled": false,
  "queue-stalled-enabled": true,
  "queue-stalled-minutes": 30,
  "rename-partial-files": true,
  "rpc-version": 15,
  "rpc-version-minimum": 1,
  "script-torrent-done-enabled": true,
  "script-torrent-done-filename": "/path/to/script",
  "seed-queue-enabled": false,
  "seed-queue-size": 10,
  "seedRatioLimit": 2,
  "seedRatioLimited": false,
  "speed-limit-down": 100,
  "speed-limit-down-enabled": false,
  "speed-limit-up": 100,
  "speed-limit-up-enabled": false,
  "start-added-torrents": true,
  "trash-original-torrent-files": false,
  "units": {
    "memory-bytes": 1024,
    "memory-units": [
      "KiB",
      "MiB",
      "GiB",
      "TiB"
    ],
    "size-bytes": 1000,
    "size-units": [
      "kB",
      "MB",
      "GB",
      "TB"
    ],
    "speed-bytes": 1000,
    "speed-units": [
      "kB/s",
      "MB/s",
      "GB/s",
      "TB/s"
    ]
  },
  "utp-enabled": true,
  "version": "2.94 (d8e60ee44f)"
}`
