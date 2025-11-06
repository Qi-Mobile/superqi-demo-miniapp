const BASE_URL = "http://172.20.10.2:1999";

class Header extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: "open" });
    }

    connectedCallback() {
        fetch("common/header.html")
            .then(res => res.text())
            .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, "text/html");
                const template = doc.querySelector("template");
                this.shadowRoot.appendChild(template.content.cloneNode(true));
                this.updateBreadcrumbs();
            });
    }

    updateBreadcrumbs() {
        const breadcrumbsContainer = this.shadowRoot.querySelector('#breadcrumbs');
        if (!breadcrumbsContainer) return;

        const path = window.location.pathname;
        const filename = path.substring(path.lastIndexOf('/') + 1) || 'index.html';

        let breadcrumbs = '<a href="index.html" class="text-gray-500 hover:text-gray-700">Home</a>';

        // Category pages
        if (filename.startsWith('category-')) {
            const categoryName = this.getCategoryName(filename);
            breadcrumbs += ` &nbsp;<span class="text-gray-400">/</span>&nbsp; <span class="text-gray-700">${categoryName}</span>`;
        }
        // Individual function pages
        else if (filename !== 'index.html') {
            const categoryFile = this.getCategoryForPage(filename);
            if (categoryFile) {
                const categoryName = this.getCategoryName(categoryFile);
                breadcrumbs += ` &nbsp;<span class="text-gray-400">/</span>&nbsp; <a href="${categoryFile}" class="text-gray-500 hover:text-gray-700">${categoryName}</a>`;
            }
            const pageName = this.getPageName(filename);
            breadcrumbs += ` &nbsp;<span class="text-gray-400">/</span>&nbsp; <span class="text-gray-700">${pageName}</span>`;
        }

        breadcrumbsContainer.innerHTML = breadcrumbs;
    }

    getCategoryName(filename) {
        const map = {
            'category-ui.html': 'UI Components',
            'category-media.html': 'Media',
            'category-storage.html': 'Storage',
            'category-auth.html': 'Authentication & Payment',
            'category-file.html': 'File',
            'category-location.html': 'Location',
            'category-network.html': 'Network',
            'category-device.html': 'Device',
        };
        return map[filename] || '';
    }

    getCategoryForPage(filename) {
        const pageCategories = {
            'Toast.html': 'category-ui.html',
            'Loading.html': 'category-ui.html',
            'confirm.html': 'category-ui.html',
            'prompt.html': 'category-ui.html',
            'showActionSheet.html': 'category-ui.html',
            'datePicker.html': 'category-ui.html',
            'MultiLevelSelect.html': 'category-ui.html',
            'choosePhoneContact.html': 'category-ui.html',
            'hideKeyboard.html': 'category-ui.html',
            'navigationBar.html': 'category-ui.html',
            'backgroundColor.html': 'category-ui.html',
            'chooseImage.html': 'category-media.html',
            'createVideoContext.html': 'category-media.html',
            'storage.html': 'category-storage.html',
            'authCode.html': 'category-auth.html',
            'saveFile.html': 'category-file.html',
            'getFileInfo.html': 'category-file.html',
            'getSavedFileList.html': 'category-file.html',
            'getSavedFileInfo.html': 'category-file.html',
            'openDocument.html': 'category-file.html',
            'getLocation.html': 'category-location.html',
            'openLocation.html': 'category-location.html',
            'request.html': 'category-network.html',
            'downloadFile.html': 'category-network.html',
            'getSystemInfo.html': 'category-device.html',
            'getNetworkType.html': 'category-network.html',
            'clipboard.html': 'category-device.html',
            'setClipboard.html': 'category-device.html',
            'vibrate.html': 'category-device.html',
            'makePhoneCall.html': 'category-device.html',
            'setKeepScreenOn.html': 'category-device.html',
            'getScreenBrightness.html': 'category-device.html',
            'setScreenBrightness.html': 'category-device.html',
            'setting.html': 'category-device.html',
            'addPhoneContact.html': 'category-device.html',
            'scan.html': 'category-device.html',
            'getBatteryInfo.html': 'category-device.html',
            'agreementPayment.html': 'category-auth.html',
            'openBrowser.html': 'category-network.html',
            'imageRelate.html': 'category-media.html',
            'removeSavedFile.html': 'category-file.html',
        };
        return pageCategories[filename] || null;
    }

    getPageName(filename) {
        const nameMap = {
            'Toast.html': 'Show Toast',
            'Loading.html': 'Show and Hide Loading',
            'alert.html': 'Alert',
            'confirm.html': 'Confirm',
            'prompt.html': 'Prompt',
            'showActionSheet.html': 'Show Action Sheet',
            'chooseImage.html': 'Choose Image',
            'saveImage.html': 'Save Image',
            'createVideoContext.html': 'Video Context',
            'datePicker.html': 'Date Picker',
            'MultiLevelSelect.html': 'Multi Level Select',
            'choosePhoneContact.html': 'Choose Phone Contact',
            'hideKeyboard.html': 'Hide Keyboard',
            'navigationBar.html': 'Navigation Bar',
            'backgroundColor.html': 'Background Color',
            'storage.html': 'Local Storage',
            'authCode.html': 'Auth Code & Payment',
            'saveFile.html': 'Save File',
            'getFileInfo.html': 'Get File Info',
            'getSavedFileList.html': 'Get Saved File List',
            'getSavedFileInfo.html': 'Get Saved File Info',
            'openDocument.html': 'Open Document',
            'getLocation.html': 'Get Location',
            'openLocation.html': 'Open Location',
            'request.html': 'Request',
            'downloadFile.html': 'Download File',
            'getSystemInfo.html': 'Get System Info',
            'getNetworkType.html': 'Get Network Type',
            'clipboard.html': 'Clipboard',
            'setClipboard.html': 'Set Clipboard',
            'vibrate.html': 'Vibrate',
            'makePhoneCall.html': 'Make Phone Call',
            'setKeepScreenOn.html': 'Set Keep Screen On',
            'getScreenBrightness.html': 'Get Screen Brightness',
            'setScreenBrightness.html': 'Set Screen Brightness',
            'setting.html': 'User Authorization Settings',
            'addPhoneContact.html': 'Add Phone Contact',
            'scan.html': 'Scan QR Code',
            'getBatteryInfo.html': 'Get Battery Info',
            'agreementPayment.html': 'Agreement Payment',
            'openBrowser.html': 'Open Browser',
            'imageRelate.html': 'Preview and Save Image',
            'removeSavedFile.html': 'Remove Saved File',
        };
        return nameMap[filename] || filename.replace('.html', '');
    }
}

customElements.define("miniapp-header", Header);

class Console extends HTMLElement {
    constructor() {
        super();
        this.attachShadow({ mode: "open" });
        this.logs = [];
        this.originalConsole = {};
        this.horizontalScrollEnabled = false;
        this.isCollapsed = false;
        this.setupConsoleInterceptor();
    }

    connectedCallback() {
        fetch("common/console.html")
            .then(res => res.text())
            .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, "text/html");
                const template = doc.querySelector("template");
                this.shadowRoot.appendChild(template.content.cloneNode(true));
                this.updateDisplay();
                this.applyInitialCollapsedState();
            });
    }

    setupConsoleInterceptor() {
        const methods = ['log', 'warn', 'error', 'info', 'debug'];
        
        methods.forEach(method => {
            this.originalConsole[method] = console[method];
            console[method] = (...args) => {
                this.addLog(method, args);
                this.originalConsole[method](...args);
            };
        });
    }

    addLog(type, args) {
        const timestamp = new Date().toLocaleTimeString();
        const message = args.map(arg => {
            if (typeof arg === 'object') {
                return JSON.stringify(arg, null, 2);
            }
            return String(arg);
        }).join(' ');

        this.logs.push({ type, message, timestamp });
        this.updateDisplay();
    }

    updateDisplay() {
        const consoleOutput = this.shadowRoot.querySelector('#console-output');
        if (consoleOutput) {
            const wrapClass = this.horizontalScrollEnabled ? '' : ' wrap-mode';
            consoleOutput.innerHTML = this.logs.map(log => 
                `<div class="console-entry console-${log.type} mb-1${wrapClass}">
<span class="console-timestamp text-gray-500 mr-2">[${log.timestamp}]</span>
<span class="console-type font-bold mr-2">[${log.type.toUpperCase()}] - <span class="console-message text-white">${log.message}</span></span>
</div>`
            ).join('');
            consoleOutput.scrollTop = consoleOutput.scrollHeight;
            
            // Update the toggle button state
            this.updateScrollMode();
        }
    }

    clear() {
        this.logs = [];
        this.updateDisplay();
    }

    toggleHorizontalScroll() {
        this.horizontalScrollEnabled = !this.horizontalScrollEnabled;
        this.updateScrollMode();
    }

    updateScrollMode() {
        const consoleEntries = this.shadowRoot.querySelectorAll('.console-entry');
        const toggleButton = this.shadowRoot.querySelector('#scroll-toggle');
        
        if (this.horizontalScrollEnabled) {
            consoleEntries.forEach(entry => entry.classList.remove('wrap-mode'));
            if (toggleButton) {
                toggleButton.textContent = 'Scroll';
                toggleButton.className = 'text-blue-400 text-sm px-2 py-1 rounded';
            }
        } else {
            consoleEntries.forEach(entry => entry.classList.add('wrap-mode'));
            if (toggleButton) {
                toggleButton.textContent = 'Wrap';
                toggleButton.className = 'text-green-400 text-sm px-2 py-1 rounded';
            }
        }
    }

    applyInitialCollapsedState() {
        const container = this.shadowRoot.querySelector('#console-container');
        const collapseButton = this.shadowRoot.querySelector('#collapse-button');
        
        if (this.isCollapsed) {
            container.classList.add('collapsed');
            collapseButton.classList.add('collapsed');
            collapseButton.title = 'Expand Console';
        }
    }

    toggleCollapse() {
        this.isCollapsed = !this.isCollapsed;
        const container = this.shadowRoot.querySelector('#console-container');
        const collapseButton = this.shadowRoot.querySelector('#collapse-button');
        
        if (this.isCollapsed) {
            container.classList.add('collapsed');
            collapseButton.classList.add('collapsed');
            collapseButton.title = 'Expand Console';
        } else {
            container.classList.remove('collapsed');
            collapseButton.classList.remove('collapsed');
            collapseButton.title = 'Collapse Console';
        }
    }
}

customElements.define("miniapp-console", Console);