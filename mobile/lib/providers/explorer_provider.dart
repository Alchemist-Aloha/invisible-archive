import 'package:flutter/material.dart';
import '../api.dart';
import '../models.dart';

class ExplorerProvider with ChangeNotifier {
  String _currentPath = '/';
  List<FileItem> _items = [];
  bool _isLoading = false;
  String? _error;
  final List<String> _history = [];

  String get currentPath => _currentPath;
  List<FileItem> get items => _items;
  bool get isLoading => _isLoading;
  String? get error => _error;
  bool get canGoBack => _history.isNotEmpty || _currentPath != '/';

  final ApiService api;

  ExplorerProvider(this.api) {
    fetchList(_currentPath);
  }

  Future<void> fetchList(String path, {bool pushToHistory = false}) async {
    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      final response = await api.fetchList(path);
      
      if (pushToHistory && _currentPath != response.effectivePath) {
        _history.add(_currentPath);
      }
      
      _items = response.items;
      _currentPath = response.effectivePath;
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  Future<void> navigateTo(String path) async {
    await fetchList(path, pushToHistory: true);
  }

  Future<void> goBack() async {
    if (_history.isNotEmpty) {
      final previousPath = _history.removeLast();
      // Use fetchList without pushing to history to avoid cycles
      await fetchList(previousPath, pushToHistory: false);
    } else if (_currentPath != '/') {
      // Fallback for when history is empty but we are not at root
      final parts = _currentPath.split('/').where((p) => p.isNotEmpty).toList();
      if (parts.length > 1) {
        parts.removeLast();
        await navigateTo('/' + parts.join('/'));
      } else {
        await navigateTo('/');
      }
      // Since navigateTo pushed to history, we might want to clear it if we want 
      // a clean "one way" back to root. But usually navigateTo is fine.
    }
  }

  Future<void> search(String query) async {
    if (query.isEmpty) {
      await fetchList(_currentPath);
      return;
    }

    _isLoading = true;
    _error = null;
    notifyListeners();

    try {
      // We don't push search to history for now, or we could if desired
      _items = await api.search(query);
    } catch (e) {
      _error = e.toString();
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }
}
