import { TestBed, fakeAsync, tick } from '@angular/core/testing';
import { ToastService } from './toast.service';

describe('ToastService', () => {
  let service: ToastService;

  beforeEach(() => {
    TestBed.configureTestingModule({});
    service = TestBed.inject(ToastService);
  });

  afterEach(() => {
    service.clear();
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  it('should start with empty toasts array', () => {
    expect(service.toasts()).toEqual([]);
  });

  it('should add a toast with show method', () => {
    service.show('Test message', 'info', 0);
    
    const toasts = service.toasts();
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Test message');
    expect(toasts[0].type).toBe('info');
  });

  it('should add a success toast', () => {
    service.success('Success message', 0);
    
    const toasts = service.toasts();
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Success message');
    expect(toasts[0].type).toBe('success');
  });

  it('should add an error toast', () => {
    service.error('Error message', 0);
    
    const toasts = service.toasts();
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Error message');
    expect(toasts[0].type).toBe('error');
  });

  it('should add an info toast', () => {
    service.info('Info message', 0);
    
    const toasts = service.toasts();
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Info message');
    expect(toasts[0].type).toBe('info');
  });

  it('should add a warning toast', () => {
    service.warning('Warning message', 0);
    
    const toasts = service.toasts();
    expect(toasts.length).toBe(1);
    expect(toasts[0].message).toBe('Warning message');
    expect(toasts[0].type).toBe('warning');
  });

  it('should assign unique IDs to toasts', () => {
    service.show('First', 'info', 0);
    service.show('Second', 'info', 0);
    service.show('Third', 'info', 0);
    
    const toasts = service.toasts();
    expect(toasts.length).toBe(3);
    expect(toasts[0].id).not.toBe(toasts[1].id);
    expect(toasts[1].id).not.toBe(toasts[2].id);
  });

  it('should remove a toast by ID', () => {
    service.show('First', 'info', 0);
    service.show('Second', 'info', 0);
    
    const toasts = service.toasts();
    const firstId = toasts[0].id;
    
    service.remove(firstId);
    
    const remainingToasts = service.toasts();
    expect(remainingToasts.length).toBe(1);
    expect(remainingToasts[0].message).toBe('Second');
  });

  it('should clear all toasts', () => {
    service.show('First', 'info', 0);
    service.show('Second', 'info', 0);
    service.show('Third', 'info', 0);
    
    expect(service.toasts().length).toBe(3);
    
    service.clear();
    
    expect(service.toasts().length).toBe(0);
  });

  it('should auto-remove toast after duration', fakeAsync(() => {
    service.show('Auto-remove', 'info', 1000);
    
    expect(service.toasts().length).toBe(1);
    
    tick(1000);
    
    expect(service.toasts().length).toBe(0);
  }));

  it('should not auto-remove toast with duration 0', fakeAsync(() => {
    service.show('Persistent', 'info', 0);
    
    expect(service.toasts().length).toBe(1);
    
    tick(10000);
    
    expect(service.toasts().length).toBe(1);
  }));

  it('should handle multiple toasts with different durations', fakeAsync(() => {
    service.show('Short', 'info', 500);
    service.show('Medium', 'info', 1500);
    service.show('Long', 'info', 3000);
    
    expect(service.toasts().length).toBe(3);
    
    tick(500);
    expect(service.toasts().length).toBe(2);
    
    tick(1000);
    expect(service.toasts().length).toBe(1);
    
    tick(1500);
    expect(service.toasts().length).toBe(0);
  }));

  it('should use default duration for success toast', fakeAsync(() => {
    service.success('Success');
    
    expect(service.toasts().length).toBe(1);
    
    tick(3000);
    
    expect(service.toasts().length).toBe(0);
  }));

  it('should use default duration for error toast', fakeAsync(() => {
    service.error('Error');
    
    expect(service.toasts().length).toBe(1);
    
    tick(5000);
    
    expect(service.toasts().length).toBe(0);
  }));
});

