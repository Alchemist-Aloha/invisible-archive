import 'dart:convert';
import 'package:http/http.dart' as http;
import 'models.dart';

class ApiService {
  final String baseUrl;

  ApiService(String baseUrl) : baseUrl = baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length - 1) : baseUrl;

  // Static field to hold the instance
  static String? _currentBaseUrl;
  static ApiService? _instance;

  static ApiService getInstance(String baseUrl) {
    if (_instance == null || _currentBaseUrl != baseUrl) {
      _currentBaseUrl = baseUrl;
      _instance = ApiService(baseUrl);
    }
    return _instance!;
  }

  String get apiBase => '$baseUrl/api';

  Future<ListResponse> fetchList(String path) async {
    final response = await http.get(Uri.parse('$apiBase/ls?path=${Uri.encodeComponent(path)}'));
    if (response.statusCode == 200) {
      return ListResponse.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to load file list: ${response.statusCode}');
    }
  }

  Future<List<FileItem>> search(String q) async {
    final response = await http.get(Uri.parse('$apiBase/search?q=${Uri.encodeComponent(q)}'));
    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((e) => FileItem.fromJson(e as Map<String, dynamic>)).toList();
    } else {
      throw Exception('Search failed: ${response.statusCode}');
    }
  }

  Future<List<FileItem>> fetchRandom(String path, {int limit = 50}) async {
    final response = await http.get(Uri.parse('$apiBase/random?path=${Uri.encodeComponent(path)}&limit=$limit'));
    if (response.statusCode == 200) {
      final List<dynamic> data = jsonDecode(response.body);
      return data.map((e) => FileItem.fromJson(e as Map<String, dynamic>)).toList();
    } else {
      throw Exception('Failed to fetch random items: ${response.statusCode}');
    }
  }

  String getRawUrl(String path, {bool download = false}) {
    final cleanPath = path.startsWith('/') ? path.substring(1) : path;
    final encodedPath = cleanPath.split('/').map(Uri.encodeComponent).join('/');
    var url = '$apiBase/raw/$encodedPath';
    if (download) {
      url += '?download=1';
    }
    return url;
  }

  String getThumbUrl(String path) {
    return '$apiBase/thumb?path=${Uri.encodeComponent(path)}';
  }

  Future<String> fetchText(String path) async {
    final url = getRawUrl(path);
    final response = await http.get(Uri.parse(url));
    if (response.statusCode == 200) {
      return response.body;
    } else {
      throw Exception('Failed to fetch text: ${response.statusCode}');
    }
  }
}
