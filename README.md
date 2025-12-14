# SuperQi Mini App Demo

Demo application showing how to integrate with SuperQi's mini app platform. Includes both Go and Node.js backend implementations, plus frontend examples for all major SuperQi Bridge APIs.

## Project Structure

- `frontend/` - HTML pages demonstrating SuperQi Bridge API functionality
- `backend-go/` - Go backend with Alipay integration
- `backend-node/` - Node.js/Express backend with Alipay integration

Both backends provide the same API endpoints. Use whichever fits your stack.

## What's Included

The frontend demo covers these API categories:

**UI Components** - Toast, loading, dialogs, action sheets, date pickers, selectors, keyboard controls, navigation bar, background colors

**Media** - Image and video handling

**Storage** - Local storage operations

**Authentication & Payment** - Auth code flow, token management, payment processing

**File** - File selection, saving, opening documents, file info

**Location** - Geolocation services

**Network** - HTTP requests, file downloads

**Device** - Device info, network type, clipboard, vibration, phone calls, screen controls, battery info, contacts

**Messages** - Notifications

## Running the Demo

### Frontend
```bash
cd frontend
npm install
npm run dev -- --host
```

### Backend

**Node.js backend:**
```bash
cd backend-node
npm install
npm run dev
```

**Go backend:**
```bash
cd backend-go
go mod download
go run main.go
```

## Contribution

We welcome contributions to the Sample Mini App! This project serves as a reference implementation for SuperQi mini app development.

### How to Contribute

1. **Fork the repository** and create your feature branch from `main`
2. **Make your changes** following the existing code style and patterns
3. **Test your changes** thoroughly:
   - For backend changes: Test API endpoints and ensure proper error handling
   - For frontend changes: Test in the SuperQi app environment
4. **Update documentation** if you're adding new features or examples
5. **Submit a pull request** with a clear description of your changes

### Areas for Contribution

- **New Examples**: Add more H5 pages demonstrating different SuperQi APIs
- **Documentation**: Improve code comments, add tutorials, or enhance README files
- **Bug Fixes**: Report and fix any issues you encounter

### Getting Help

If you have questions about contributing, feel free to:
- Open an issue for discussion
- Check the [SuperQi Developers Guide](https://superqi-dev-docs.pages.dev/) for platform-specific questions
- Review existing code examples for patterns and best practices


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Links

- [SuperQi Developers Guide](https://superqi.qi-mobile.tech/)
- [SuperQI Miniapps Console](https://miniapps.qi.iq/gotoconsole)

