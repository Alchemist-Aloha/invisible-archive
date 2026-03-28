import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:mocktail/mocktail.dart';
import 'package:invisible_archive_mobile/api.dart';
import 'package:invisible_archive_mobile/models.dart';

class MockClient extends Mock implements http.Client {}

void main() {
  group('ApiService Tests', () {
    late ApiService apiService;
    late MockClient mockClient;

    setUpAll(() {
      registerFallbackValue(Uri.parse('http://test.com'));
    });

    setUp(() {
      mockClient = MockClient();
      apiService = ApiService('http://test.com/', client: mockClient);
    });

    group('URL Generation', () {
      test('baseUrl trimming trailing slash', () {
        expect(apiService.baseUrl, 'http://test.com');
      });

      test('apiBase getter', () {
        expect(apiService.apiBase, 'http://test.com/api');
      });

      test('getRawUrl', () {
        expect(apiService.getRawUrl('test/path'), 'http://test.com/api/raw/test/path');
        expect(apiService.getRawUrl('/test/path', download: true), 'http://test.com/api/raw/test/path?download=1');
        expect(apiService.getRawUrl('file with spaces'), 'http://test.com/api/raw/file%20with%20spaces');
      });

      test('getThumbUrl', () {
        expect(apiService.getThumbUrl('test/path'), 'http://test.com/api/thumb?path=test%2Fpath');
      });
    });

    group('fetchList', () {
      test('returns ListResponse on 200', () async {
        final mockJson = {
          'effective_path': 'test',
          'items': [
            {'name': 'file1.txt', 'path': 'test/file1.txt', 'is_dir': false, 'size': 100, 'mod_time': 123456789, 'capabilities': 1}
          ]
        };
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response(jsonEncode(mockJson), 200));

        final response = await apiService.fetchList('test');
        expect(response.effectivePath, 'test');
        expect(response.items.length, 1);
        expect(response.items.first.name, 'file1.txt');
        verify(() => mockClient.get(Uri.parse('http://test.com/api/ls?path=test'))).called(1);
      });

      test('appends sort and order parameters', () async {
        final mockJson = {'effective_path': 'test', 'items': []};
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response(jsonEncode(mockJson), 200));

        await apiService.fetchList('test', sort: 'name', order: 'desc');
        verify(() => mockClient.get(Uri.parse('http://test.com/api/ls?path=test&sort=name&order=desc'))).called(1);
      });

      test('throws Exception on non-200', () async {
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response('Error', 404));
        expect(() => apiService.fetchList('test'), throwsException);
      });
    });

    group('search', () {
      test('returns List<FileItem> on 200', () async {
        final mockJson = [
          {'name': 'file1.txt', 'path': 'test/file1.txt', 'is_dir': false, 'size': 100, 'mod_time': 123456789, 'capabilities': 1}
        ];
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response(jsonEncode(mockJson), 200));

        final response = await apiService.search('query');
        expect(response.length, 1);
        expect(response.first.name, 'file1.txt');
        verify(() => mockClient.get(Uri.parse('http://test.com/api/search?q=query'))).called(1);
      });

      test('throws Exception on non-200', () async {
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response('Error', 500));
        expect(() => apiService.search('query'), throwsException);
      });
    });

    group('fetchRandom', () {
      test('returns List<FileItem> on 200', () async {
        final mockJson = [
          {'name': 'file1.txt', 'path': 'test/file1.txt', 'is_dir': false, 'size': 100, 'mod_time': 123456789, 'capabilities': 1}
        ];
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response(jsonEncode(mockJson), 200));

        final response = await apiService.fetchRandom('test', limit: 10);
        expect(response.length, 1);
        expect(response.first.name, 'file1.txt');
        verify(() => mockClient.get(Uri.parse('http://test.com/api/random?path=test&limit=10'))).called(1);
      });

      test('throws Exception on non-200', () async {
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response('Error', 403));
        expect(() => apiService.fetchRandom('test'), throwsException);
      });
    });

    group('fetchText', () {
      test('returns text string on 200', () async {
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response('Hello World', 200));

        final response = await apiService.fetchText('test.txt');
        expect(response, 'Hello World');
        verify(() => mockClient.get(Uri.parse('http://test.com/api/raw/test.txt'))).called(1);
      });

      test('throws Exception on non-200', () async {
        when(() => mockClient.get(any())).thenAnswer((_) async => http.Response('Error', 404));
        expect(() => apiService.fetchText('test.txt'), throwsException);
      });
    });
  });
}
