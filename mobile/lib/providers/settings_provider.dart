import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class SettingsProvider with ChangeNotifier {
  String _serverUrl = 'http://10.0.2.2:8080'; // Default for Android Emulator
  bool _isDarkMode = false;
  String _layoutMode = 'grid';

  String get serverUrl => _serverUrl;
  bool get isDarkMode => _isDarkMode;
  String get layoutMode => _layoutMode;

  SettingsProvider() {
    _loadSettings();
  }

  Future<void> _loadSettings() async {
    final prefs = await SharedPreferences.getInstance();
    _serverUrl = prefs.getString('serverUrl') ?? _serverUrl;
    _isDarkMode = prefs.getBool('isDarkMode') ?? _isDarkMode;
    _layoutMode = prefs.getString('layoutMode') ?? _layoutMode;
    notifyListeners();
  }

  Future<void> setServerUrl(String url) async {
    _serverUrl = url;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('serverUrl', url);
    notifyListeners();
  }

  Future<void> setDarkMode(bool value) async {
    _isDarkMode = value;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setBool('isDarkMode', value);
    notifyListeners();
  }

  Future<void> setLayoutMode(String mode) async {
    _layoutMode = mode;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('layoutMode', mode);
    notifyListeners();
  }
}
