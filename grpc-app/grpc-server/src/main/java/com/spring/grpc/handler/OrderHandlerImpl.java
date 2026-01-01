package com.spring.grpc.handler;

import com.spring.grpc.dto.OrderInfo;
import com.spring.grpc.dto.OrderStatus;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
public class OrderHandlerImpl implements OrderHandler{

    private final OrderInfoService orderInfoService;
    private final OrderStatusService orderStatusService;
    public OrderHandlerImpl(OrderInfoService orderInfoService, OrderStatusService orderStatusService) {
        this.orderInfoService = orderInfoService;
        this.orderStatusService = orderStatusService;
    }

    @Override
    public List<OrderInfo> findOrderInfoForUser(String userName) {
        return orderInfoService.findByUserName(userName);
    }

    @Override
    public OrderInfo addOrderInfo(OrderInfo orderInfo) {
        return orderInfoService.save(orderInfo);
    }

    @Override
    public List<OrderStatus> getOrderStatusByUserNameOrOrderId(String userName, long orderId) {
       return orderStatusService.findByUserNameOrOrderId(userName,orderId);
    }

    @Override
    public OrderInfo findOrderInfoByUserNameAndOrderId(String userName, long orderId) {
        return orderInfoService.findByUserNameAndOrderId(userName,orderId);
    }

    @Override
    public OrderStatus addOrderStatus(OrderStatus status) {
        return orderStatusService.save(status);
    }

    @Override
    public OrderInfo updateOrderInfo(OrderInfo info){
        return orderInfoService.updateOrInsert(info);
    }

}
